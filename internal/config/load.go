package config

import (
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var logger = pkg.Logger.With("component", "config")

func New() *Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		logger.Warn("No ..env file found, using system environment")
	}

	if err := env.Parse(&cfg); err != nil {
		logger.Error("error in parse env", "error", err)
	}

	return &cfg
}
