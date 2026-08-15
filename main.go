package main

import (
	"context"
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	RedisUrl string `env:"REDIS_URL" envDefault:"redis://:@localhost:6379/"`
}

func main() {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := initRedis(context.Background(), cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
	}
	err = initWeb(redisClient)
	if err != nil {
		log.Fatal(err)
	}
}
