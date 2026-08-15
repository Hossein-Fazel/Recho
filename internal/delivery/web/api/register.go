package api

import (
	"github.com/Hossein-Fazel/Recho/internal/usecase"
	"github.com/labstack/echo/v4"
)

func Register(g *echo.Group, auth usecase.AuthService, access usecase.AccessToken) {
	authGroup := g.Group("/auth")
	authHandler := NewAuthHandler(auth, access)
	authHandler.RegisterRoutes(authGroup)

}