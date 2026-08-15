package cmd

import (
	"context"

	postgres_repo "github.com/Hossein-Fazel/Recho/internal/adapter/postgres"
	"github.com/Hossein-Fazel/Recho/internal/config"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	"github.com/Hossein-Fazel/Recho/internal/infra/auth"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
	"github.com/Hossein-Fazel/Recho/internal/usecase"

	"github.com/Hossein-Fazel/Recho/pkg"
)

var logger = pkg.Logger.With("component", "app")

func Run() {
	ctx := context.Background()
	conf := config.New()

	logger.Info("Initializing database")
	queries := db.NewQueries(ctx, conf.DB)

	logger.Info("Initializing repos")
	authRepo := postgres_repo.NewRefreshTokenRepo(queries)
	userRepo := postgres_repo.NewUserRepo(queries)

	logger.Info("Initializing services")
	accessToken := auth.NewJWTService(conf.Token)
	refreshToken := auth.NewRefreshTokenService(conf.Token)

	authService := usecase.NewAuthService(userRepo, accessToken, refreshToken, authRepo)

	web.Start(authService, accessToken, conf.Server)
}
