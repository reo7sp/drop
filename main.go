package main

import (
	"github.com/caarlos0/env"
	"log"
)

type config struct {
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDb       int `env:"REDIS_DB" endDefault:"0"`
}

func main() {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDb)
	if err != nil {
		log.Fatal(err)
	}
	InitWeb(redisClient)
}
