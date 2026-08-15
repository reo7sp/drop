package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	dropLifetime = time.Minute
	maxDrops     = 100
)

type repository interface {
	FetchAll(context.Context) ([]string, error)
	Save(context.Context, string) error
}

type redisRepository struct {
	client *redis.Client
}

func initRepository(ctx context.Context, redisUrl string) (*redisRepository, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	err = client.Ping(ctx).Err()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &redisRepository{client: client}, nil
}

func (r redisRepository) FetchAll(ctx context.Context) ([]string, error) {
	err := r.cleanup(ctx, time.Now())
	if err != nil {
		return nil, fmt.Errorf("clean up drops: %w", err)
	}

	drops, err := r.client.LRange(ctx, "drops", 0, maxDrops-1).Result()
	if err != nil {
		return nil, fmt.Errorf("fetch drops: %w", err)
	}

	return drops, nil
}

func (r redisRepository) cleanup(ctx context.Context, now time.Time) error {
	for {
		dropTimeStr, err := r.client.LIndex(ctx, "drop_times", -1).Result()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("fetch oldest drop timestamp: %w", err)
		}

		dropTime, err := strconv.ParseInt(dropTimeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("parse drop timestamp %q: %w", dropTimeStr, err)
		}

		if now.Sub(time.Unix(dropTime, 0)) <= dropLifetime {
			return nil
		}

		_, err = r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.RPop(ctx, "drop_times")
			pipe.RPop(ctx, "drops")
			return nil
		})
		if err != nil {
			return fmt.Errorf("remove expired drop: %w", err)
		}
	}
}

func (r redisRepository) Save(ctx context.Context, drop string) error {
	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LPush(ctx, "drops", drop)
		pipe.LPush(ctx, "drop_times", time.Now().Unix())
		pipe.LTrim(ctx, "drops", 0, maxDrops-1)
		pipe.LTrim(ctx, "drop_times", 0, maxDrops-1)
		return nil
	})
	if err != nil {
		return fmt.Errorf("save drop: %w", err)
	}

	return nil
}
