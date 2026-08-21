package api

import (
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group, auth application.AuthService, access application.AccessToken) {
	authGroup := g.Group("/auth")
	authHandler := NewAuthHandler(auth, access)
	authHandler.RegisterRoutes(authGroup)

}
