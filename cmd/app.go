package cmd

import (
	"context"

	postgres_repo "github.com/Hossein-Fazel/Recho/internal/adapter/postgres"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/config"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/middleware"
	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket"
	db "github.com/Hossein-Fazel/Recho/internal/infra/postgres"
	"github.com/Hossein-Fazel/Recho/internal/infra/token"

	"github.com/Hossein-Fazel/Recho/pkg"
)

func Run() {
	ctx := context.Background()
	conf := config.New()

	pkg.Logger.Info().Msg("Initializing database")
	database, err := db.NewDBPool(ctx, conf.DB)
	if err != nil {
		pkg.Logger.Fatal().
			AnErr("error", err).
			Msg("error in creating db")
	}

	err = db.RunMigrations(conf.DB)
	if err != nil {
		pkg.Logger.Error().
			AnErr("error", err).
			Msg("error in running migrations")
	}

	queries := db.NewQueries(database)

	pkg.Logger.Info().Msg("Initializing repos")
	authRepo := postgres_repo.NewRefreshTokenRepo(queries, database)
	userRepo := postgres_repo.NewUserRepo(queries)
	convRepo := postgres_repo.NewConversationRepo(queries, database)
	msgRepo := postgres_repo.NewMessageRepo(queries)

	pkg.Logger.Info().Msg("Initializing services")
	accessToken := token.NewJWTService(conf.Token)
	refreshToken := token.NewRefreshTokenService(conf.Token)

	authService := application.NewAuthService(userRepo, accessToken, refreshToken, authRepo)
	convService := application.NewConversationService(convRepo)
	userService := application.NewUserService(userRepo)
	msgService := application.NewMessageService(msgRepo, convRepo)

	pkg.Logger.Info().Msg("initialize websocket")
	hub := websocket.NewHub()
	go hub.Run()
	msgDelivery := application.NewMessageDelivery(msgService, convService, hub)
	ws := websocket.NewWSHandler(hub, msgDelivery)

	pkg.Logger.Info().Msg("Starting server")

	web := web.Init(web.Services{
		Auth:         authService,
		AccessToken:  accessToken,
		Conversation: convService,
		User:         userService,
	}, conf.Server)

	wsGroup := web.Group("/ws", middleware.AccessMiddleware(accessToken))
	ws.RegsiterRoutes(wsGroup)
	web.Logger.Fatal(web.Start(":" + conf.Server.Port))
}
