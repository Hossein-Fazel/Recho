package config

import (
	"github.com/Hossein-Fazel/Recho/internal/infra/auth"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
)

type Config struct {
	DB    db.Config `envPrefix:"POSTGRES_"`
	Token auth.Config `envPrefix:"TOKEN_"`
}
