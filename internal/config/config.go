package config

import (
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
	"github.com/Hossein-Fazel/Recho/internal/infra/token"
)

type Config struct {
	DB     db.Config    `envPrefix:"POSTGRES_"`
	Token  token.Config `envPrefix:"TOKEN_"`
	Server web.Config   `envPrefix:"SERVER_"`
}
