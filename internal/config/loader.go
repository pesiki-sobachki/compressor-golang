package config

import (
	"context"
	"log"

	"github.com/shanth1/gotools/conf"
)

func Load(ctx context.Context) (Config, error) {
	var cfg Config

	//read from config.yaml
	if err := conf.Load("config.yaml", &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func MustLoad(ctx context.Context) Config {
	cfg, err := Load(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}
