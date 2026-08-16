package config

import (
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func New() *Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		pkg.Logger.Warn("No ..env file found, using system environment")
	}

	if err := env.Parse(&cfg); err != nil {
		pkg.Logger.Warn("error in parse env", "error", err)
	}

	return &cfg
}
