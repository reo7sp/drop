package main

import (
	"github.com/go-redis/redis"
	"time"
	"strconv"
)

func InitRedis(redisAddr string, redisPassword string, redisDb int) (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDb,
	})
	err = client.Ping().Err()
	return
}

func FetchAllDrops(redisClient *redis.Client) (drops []string, err error) {
	err = CleanupDrops(redisClient)
	if err != nil {
		return
	}

	drops, err = redisClient.LRange("drops", 0, -1).Result()
	return
}

func CleanupDrops(redisClient *redis.Client) (error) {
	const fiveMinutes = 5 * 60
	now := time.Now().Unix()

	for {
		ok, err := redisClient.Exists("drop_times").Result()
		if err != nil {
			return err
		}
		if ok == 0 {
			return nil
		}

		dropTimeStr, err := redisClient.LIndex("drop_times", -1).Result()
		if err != nil {
			return err
		}
		dropTime, err := strconv.ParseInt(dropTimeStr, 10, 64)
		if err != nil {
			return err
		}

		if now - dropTime > fiveMinutes {
			err = redisClient.RPop("drop_times").Err()
			if err != nil {
				return err
			}
			err = redisClient.RPop("drops").Err()
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
	err = redisClient.LPush("drops", drop).Err()
	if err != nil {
		return
	}
	err = redisClient.LPush("drop_times", time.Now().Unix()).Err()
	return
}
