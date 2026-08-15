package main

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const dropLifetime = time.Minute

func initRedis(ctx context.Context, redisUrl string) (client *redis.Client, err error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return
	}
	client = redis.NewClient(opts)
	err = client.Ping(ctx).Err()
	if err != nil {
		_ = client.Close()
		client = nil
	}
	return
}

func fetchAllDrops(ctx context.Context, redisClient *redis.Client) (drops []string, err error) {
	err = cleanupDrops(ctx, redisClient, time.Now())
	if err != nil {
		return
	}

	drops, err = redisClient.LRange(ctx, "drops", 0, -1).Result()
	return
}

func cleanupDrops(ctx context.Context, redisClient *redis.Client, now time.Time) error {
	for {
		dropTimeStr, err := redisClient.LIndex(ctx, "drop_times", -1).Result()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			return err
		}
		dropTime, err := strconv.ParseInt(dropTimeStr, 10, 64)
		if err != nil {
			return err
		}

		if now.Sub(time.Unix(dropTime, 0)) > dropLifetime {
			_, err = redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.RPop(ctx, "drop_times")
				pipe.RPop(ctx, "drops")
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}

func saveDrop(ctx context.Context, redisClient *redis.Client, drop string) (err error) {
	_, err = redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LPush(ctx, "drops", drop)
		pipe.LPush(ctx, "drop_times", time.Now().Unix())
		return nil
	})
	return
}
