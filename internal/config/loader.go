package config

import (
	"context"
	"log"

	"os"

	"gopkg.in/yaml.v3"
)

func Load(ctx context.Context) (Config, error) {
	var cfg Config

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return Config{}, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
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
