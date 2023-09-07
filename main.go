package main

import (
	"github.com/caarlos0/env"
	"log"
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

	redisClient, err := initRedis(cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
	}
	initWeb(redisClient)
}
