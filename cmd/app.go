package cmd

import (
	"context"

	postgres_repo "github.com/Hossein-Fazel/Recho/internal/adapter/postgres"
	"github.com/Hossein-Fazel/Recho/internal/config"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
	"github.com/Hossein-Fazel/Recho/internal/infra/token"
	"github.com/Hossein-Fazel/Recho/internal/usecase"

	"github.com/Hossein-Fazel/Recho/pkg"
)

func Run() {
	ctx := context.Background()
	conf := config.New()

	pkg.Logger.Info().Msg("Initializing database")
	database, err := db.NewDBTX(ctx, conf.DB)
	if err != nil {
		pkg.Logger.Error().
			AnErr("error", err).
			Msg("error in creating db")
	}
	queries := db.NewQueries(ctx, database, conf.DB)

	pkg.Logger.Info().Msg("Initializing repos")
	authRepo := postgres_repo.NewRefreshTokenRepo(queries, database)
	userRepo := postgres_repo.NewUserRepo(queries)

	pkg.Logger.Info().Msg("Initializing services")
	accessToken := token.NewJWTService(conf.Token)
	refreshToken := token.NewRefreshTokenService(conf.Token)

	authService := usecase.NewAuthService(userRepo, accessToken, refreshToken, authRepo)

	web.Start(authService, accessToken, conf.Server)
}
