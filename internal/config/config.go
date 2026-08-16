package config

import (
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	"github.com/Hossein-Fazel/Recho/internal/infra/token"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
)

type Config struct {
	DB     db.Config   `envPrefix:"POSTGRES_"`
	Token  token.Config `envPrefix:"TOKEN_"`
	Server web.Config  `envPrefix:"SERVER_"`
}
