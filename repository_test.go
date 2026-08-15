package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}

func TestInitRedis(t *testing.T) {
	server := miniredis.RunT(t)

	r, err := initRepository(context.Background(), "redis://"+server.Addr()+"/0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.client.Close() })

	require.NoError(t, r.client.Ping(context.Background()).Err())
}

func TestInitRedisRejectsInvalidURL(t *testing.T) {
	_, err := initRepository(context.Background(), "://invalid")
	require.Error(t, err)
}

func TestSaveAndFetchDrops(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	r := redisRepository{client: client}

	for _, drop := range []string{"first", "second"} {
		require.NoError(t, r.Save(ctx, drop))
	}

	drops, err := r.FetchAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"second", "first"}, drops)
}

func TestRepositoryLimitsDrops(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	r := redisRepository{client: client}

	for i := 0; i <= maxDrops; i++ {
		require.NoError(t, r.Save(ctx, fmt.Sprintf("drop-%d", i)))
	}

	drops, err := r.FetchAll(ctx)
	require.NoError(t, err)
	require.Len(t, drops, maxDrops)
	assert.Equal(t, "drop-100", drops[0])
	assert.Equal(t, "drop-1", drops[maxDrops-1])

	timesCount, err := client.LLen(ctx, "drop_times").Result()
	require.NoError(t, err)
	assert.EqualValues(t, maxDrops, timesCount)

	expiredAt := time.Now().Add(-dropLifetime - time.Second).Unix()
	require.NoError(t, client.LSet(ctx, "drop_times", -1, expiredAt).Err())

	drops, err = r.FetchAll(ctx)
	require.NoError(t, err)
	require.Len(t, drops, maxDrops-1)
	assert.Equal(t, "drop-2", drops[len(drops)-1])

	timesCount, err = client.LLen(ctx, "drop_times").Result()
	require.NoError(t, err)
	assert.EqualValues(t, maxDrops-1, timesCount)
}

func TestCleanupDropsRemovesOnlyExpiredEntries(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)
	r := redisRepository{client: client}

	require.NoError(t, client.RPush(ctx, "drops", "fresh", "boundary", "expired").Err())
	require.NoError(t, client.RPush(ctx, "drop_times",
		now.Add(-time.Minute).Unix(),
		now.Add(-dropLifetime).Unix(),
		now.Add(-dropLifetime-time.Second).Unix(),
	).Err())

	require.NoError(t, r.cleanup(ctx, now))

	drops, err := client.LRange(ctx, "drops", 0, -1).Result()
	require.NoError(t, err)
	assert.Equal(t, []string{"fresh", "boundary"}, drops)
}

func TestCleanupDropsReportsMalformedTimestamp(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	r := redisRepository{client: client}
	require.NoError(t, client.LPush(ctx, "drop_times", "not-a-timestamp").Err())

	require.Error(t, r.cleanup(ctx, time.Now()))
}

func TestStorageHonorsCanceledContext(t *testing.T) {
	server, client := testRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := redisRepository{client: client}

	assert.ErrorIs(t, r.Save(ctx, "drop"), context.Canceled)
	_, err := r.FetchAll(ctx)
	assert.ErrorIs(t, err, context.Canceled)

	canceledRepository, err := initRepository(ctx, "redis://"+server.Addr()+"/0")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, canceledRepository)
}
