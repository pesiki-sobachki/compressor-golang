package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/shanth1/gotools/conf"
	"github.com/shanth1/gotools/consts"
	"github.com/shanth1/gotools/env"
	"github.com/shanth1/gotools/flags"
)

type bootstrapConfig struct {
	AppEnv     string `flag:"env" usage:"Environment: local, dev, stage, prod"`
	ConfigPath string `flag:"config" usage:"Path to the YAML config file"`
	EnvPath    string `flag:"env-path" usage:"Path to the env file"`
}

func Load(ctx context.Context) (*Config, error) {
	boot := &bootstrapConfig{}
	if err := flags.RegisterFromStruct(boot); err != nil {
		return nil, fmt.Errorf("register flags: %w", err)
	}
	flag.Parse()

	appEnv := boot.AppEnv
	if appEnv == "" {
		if envVar, exists := os.LookupEnv("APP_ENV"); exists {
			appEnv = envVar
		} else {
			appEnv = "local"
		}
	}

	if boot.ConfigPath == "" {
		boot.ConfigPath = filepath.Join("internal/config", fmt.Sprintf("config.%s.yaml", appEnv))
	}

	cfg := &Config{}

	if _, err := os.Stat(boot.ConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at: %s", boot.ConfigPath)
	}

	if err := conf.Load(boot.ConfigPath, cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := env.LoadIntoStruct(boot.EnvPath, cfg); err != nil {
		log.Printf("Warning: failed to load env: %v", err)
	}

	cfg.Env = consts.Env(appEnv)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	return cfg, nil
}

func MustLoad(ctx context.Context) *Config {
	cfg, err := Load(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}
