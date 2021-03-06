package main

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func InitRedis(redisUrl string) (client *redis.Client, err error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return
	}
	client = redis.NewClient(opts)
	err = client.Ping(context.Background()).Err()
	return
}

func FetchAllDrops(redisClient *redis.Client) (drops []string, err error) {
	err = CleanupDrops(redisClient)
	if err != nil {
		return
	}

	drops, err = redisClient.LRange(context.Background(), "drops", 0, -1).Result()
	return
}

func CleanupDrops(redisClient *redis.Client) (error) {
	const fiveMinutes = 5 * 60
	now := time.Now().Unix()

	for {
		ok, err := redisClient.Exists(context.Background(), "drop_times").Result()
		if err != nil {
			return err
		}
		if ok == 0 {
			return nil
		}

		dropTimeStr, err := redisClient.LIndex(context.Background(), "drop_times", -1).Result()
		if err != nil {
			return err
		}
		dropTime, err := strconv.ParseInt(dropTimeStr, 10, 64)
		if err != nil {
			return err
		}

		if now - dropTime > fiveMinutes {
			err = redisClient.RPop(context.Background(), "drop_times").Err()
			if err != nil {
				return err
			}
			err = redisClient.RPop(context.Background(), "drops").Err()
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func SaveDrop(redisClient *redis.Client, drop string) (err error) {
	err = redisClient.LPush(context.Background(), "drops", drop).Err()
	if err != nil {
		return
	}
	err = redisClient.LPush(context.Background(), "drop_times", time.Now().Unix()).Err()
	return
}

