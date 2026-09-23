package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/adapter/memory"
	postgres_repo "github.com/Hossein-Fazel/Recho/internal/adapter/postgres"
	"github.com/Hossein-Fazel/Recho/internal/adapter/storage"
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
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	conf := config.New()

	// -------------------------------------------------------------------------
	// Database
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing database")

	database, err := db.NewDBPool(ctx, conf.DB)
	if err != nil {
		pkg.Logger.Fatal().
			AnErr("error", err).
			Msg("error in creating db")
	}
	defer database.Close()

	err = db.RunMigrations(conf.DB)
	if err != nil {
		pkg.Logger.Fatal().
			AnErr("error", err).
			Msg("error in running migrations")
	}

	queries := db.NewQueries(database)

	// -------------------------------------------------------------------------
	// Repositories
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing repos")

	authRepo := postgres_repo.NewRefreshTokenRepo(queries, database)
	userRepo := postgres_repo.NewUserRepo(queries)
	convRepo := postgres_repo.NewConversationRepo(queries, database)
	groupRepo := postgres_repo.NewGroupRepo(queries, database)
	msgRepo := postgres_repo.NewMessageRepo(queries, database)
	presenceRepo := memory.NewPresenceRepo()

	// -------------------------------------------------------------------------
	// Sender
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing sender")

	sender := websocket.NewSender(ctx)

	// -------------------------------------------------------------------------
	// Storage
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing storage")

	store, err := storage.New(ctx, conf.Storage)
	if err != nil {
		pkg.Logger.Fatal().
			AnErr("error", err).
			Msg("error in creating storage")
	}

	// -------------------------------------------------------------------------
	// Services
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing services")

	accessToken := token.NewJWTService(conf.Token)
	refreshToken := token.NewRefreshTokenService(conf.Token)

	storageService := application.NewStorageService(store)

	authService := application.NewAuthService(
		userRepo,
		accessToken,
		refreshToken,
		authRepo,
	)

	convService := application.NewConversationService(
		convRepo,
		storageService,
	)

	userService := application.NewUserService(
		userRepo,
		storageService,
	)

	groupService := application.NewGroupService(
		groupRepo,
		convRepo,
		storageService,
		sender,
	)

	msgService := application.NewMessageService(
		msgRepo,
		convRepo,
	)

	msgDelivery := application.NewMessageDelivery(
		msgService,
		convService,
		sender,
	)

	presenceService := application.NewPresenceService(presenceRepo, userRepo, sender)

	// -------------------------------------------------------------------------
	// WebSocket
	// -------------------------------------------------------------------------

	pkg.Logger.Info().Msg("Initializing websocket")

	hub := websocket.NewHub(ctx, sender.Channel(), presenceService)
	go hub.Run()

	
	// -------------------------------------------------------------------------
	// HTTP Server
	// -------------------------------------------------------------------------
	
	pkg.Logger.Info().Msg("Starting server")
	
	ws := websocket.NewWSHandler(hub, msgDelivery)

	e := web.Init(web.Services{
		Auth:         authService,
		AccessToken:  accessToken,
		Conversation: convService,
		User:         userService,
		Group:        groupService,
		Presence:     presenceService,
	}, conf.Server, conf.Storage)

	wsGroup := e.Group(
		"/ws",
		middleware.AccessMiddleware(accessToken),
	)

	ws.RegsiterRoutes(wsGroup)

	go func() {
		if err := e.Start(":" + conf.Server.Port); err != nil {
			if err.Error() != "http: Server closed" {
				pkg.Logger.Error().
					AnErr("error", err).
					Msg("server stopped unexpectedly")
			}
		}
	}()

	pkg.Logger.Info().
		Str("port", conf.Server.Port).
		Msg("server started")

	// -------------------------------------------------------------------------
	// Graceful Shutdown
	// -------------------------------------------------------------------------

	<-ctx.Done()

	pkg.Logger.Info().Msg("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pkg.Logger.Info().Msg("Shutting down HTTP server")

	if err := e.Shutdown(shutdownCtx); err != nil {
		pkg.Logger.Error().
			AnErr("error", err).
			Msg("error shutting down HTTP server")
	}

	pkg.Logger.Info().Msg("Shutting down websocket hub")

	if err := hub.Shutdown(shutdownCtx); err != nil {
		pkg.Logger.Error().
			AnErr("error", err).
			Msg("error shutting down websocket hub")
	}

	pkg.Logger.Info().Msg("Closing storage")

	if closer, ok := store.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			pkg.Logger.Error().
				AnErr("error", err).
				Msg("error closing storage")
		}
	}

	pkg.Logger.Info().Msg("Closing database")

	database.Close()

	pkg.Logger.Info().Msg("Shutdown complete")
}
