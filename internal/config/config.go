package config

import (
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
)

type Config struct {
	DB db.Config `envPrefix:"POSTGRES_"`
}
