package main

import (
	"context"
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

	client, err := initRedis(context.Background(), "redis://"+server.Addr()+"/0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	require.NoError(t, client.Ping(context.Background()).Err())
}

func TestInitRedisRejectsInvalidURL(t *testing.T) {
	_, err := initRedis(context.Background(), "://invalid")
	require.Error(t, err)
}

func TestSaveAndFetchDrops(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()

	for _, drop := range []string{"first", "second"} {
		require.NoError(t, saveDrop(ctx, client, drop))
	}

	drops, err := fetchAllDrops(ctx, client)
	require.NoError(t, err)
	assert.Equal(t, []string{"second", "first"}, drops)
}

func TestCleanupDropsRemovesOnlyExpiredEntries(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)

	require.NoError(t, client.RPush(ctx, "drops", "fresh", "boundary", "expired").Err())
	require.NoError(t, client.RPush(ctx, "drop_times",
		now.Add(-time.Minute).Unix(),
		now.Add(-dropLifetime).Unix(),
		now.Add(-dropLifetime-time.Second).Unix(),
	).Err())

	require.NoError(t, cleanupDrops(ctx, client, now))

	drops, err := client.LRange(ctx, "drops", 0, -1).Result()
	require.NoError(t, err)
	assert.Equal(t, []string{"fresh", "boundary"}, drops)
}

func TestCleanupDropsReportsMalformedTimestamp(t *testing.T) {
	_, client := testRedis(t)
	ctx := context.Background()
	require.NoError(t, client.LPush(ctx, "drop_times", "not-a-timestamp").Err())

	require.Error(t, cleanupDrops(ctx, client, time.Now()))
}

func TestStorageReportsRedisFailure(t *testing.T) {
	server, client := testRedis(t)
	server.Close()
	ctx := context.Background()

	assert.Error(t, saveDrop(ctx, client, "drop"))
	_, err := fetchAllDrops(ctx, client)
	assert.Error(t, err)
}

func TestStorageHonorsCanceledContext(t *testing.T) {
	server, client := testRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.ErrorIs(t, saveDrop(ctx, client, "drop"), context.Canceled)
	_, err := fetchAllDrops(ctx, client)
	assert.ErrorIs(t, err, context.Canceled)

	canceledClient, err := initRedis(ctx, "redis://"+server.Addr()+"/0")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, canceledClient)
}
