package cmd

import (
	"context"

	postgres_repo "github.com/Hossein-Fazel/Recho/internal/adapter/postgres"
	"github.com/Hossein-Fazel/Recho/internal/config"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	"github.com/Hossein-Fazel/Recho/internal/infra/token"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
	"github.com/Hossein-Fazel/Recho/internal/usecase"

	"github.com/Hossein-Fazel/Recho/pkg"
)

func Run() {
	ctx := context.Background()
	conf := config.New()

	pkg.Logger.Info("Initializing database")
	queries := db.NewQueries(ctx, conf.DB)

	pkg.Logger.Info("Initializing repos")
	authRepo := postgres_repo.NewRefreshTokenRepo(queries)
	userRepo := postgres_repo.NewUserRepo(queries)

	pkg.Logger.Info("Initializing services")
	accessToken := token.NewJWTService(conf.Token)
	refreshToken := token.NewRefreshTokenService(conf.Token)

	authService := usecase.NewAuthService(userRepo, accessToken, refreshToken, authRepo)

	web.Start(authService, accessToken, conf.Server)
}
