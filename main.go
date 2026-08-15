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
		log.Fatalf("parse config: %v", err)
	}

	r, err := initRepository(context.Background(), cfg.RedisUrl)
	if err != nil {
		log.Fatalf("initialize redis: %v", err)
	}

	err = initHandlers(r)
	if err != nil {
		log.Fatalf("start HTTP server: %v", err)
	}
}
