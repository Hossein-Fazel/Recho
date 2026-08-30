package web

import (
	"fmt"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/api"
	rechoMiddleware "github.com/Hossein-Fazel/Recho/internal/delivery/web/middleware"

	_ "github.com/Hossein-Fazel/Recho/docs"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Config struct {
	Port         string `env:"PORT"`
	SwaggerRoute string `env:"SWAGGER_ROUTE"`
	Mode         string `env:"MODE" envDefault:"development"`
}

func (c *Config) isDebug() bool {
	return c.Mode != "production" && c.Mode != "prod"
}

type Services struct {
	Auth         *application.AuthService
	AccessToken  application.AccessToken
	Conversation *application.ConverasionService
	User         *application.UserService
}

// @title Recho
// @version 0.0.0
// @description Realtime chat application
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func Init(svcs Services, conf Config) *echo.Echo {
	pkg.Logger.Info().Msg("Initializing server")

	webServer := echo.New()
	webServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	webServer.Debug = conf.isDebug()

	webServer.HTTPErrorHandler = HTTPErrorHandler

	webServer.Use(middleware.RequestLoggerWithConfig(
		middleware.RequestLoggerConfig{
			LogStatus:  true,
			LogURI:     true,
			LogMethod:  true,
			LogLatency: true,

			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

				pkg.Logger.Info().
					Str("method", v.Method).
					Str("uri", v.URI).
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Msg("http method")

				return nil
			},
		},
	))

	if webServer.Debug {
		swaggerRoute := fmt.Sprintf("/%s/*", conf.SwaggerRoute)
		webServer.GET(swaggerRoute, echoSwagger.WrapHandler)
	}

	apiGroup := webServer.Group("/api")

	accessMidleware := rechoMiddleware.AccessMiddleware(svcs.AccessToken)

	// Register handlers
	authGroup := apiGroup.Group("/auth")
	authHandler := api.NewAuthHandler(svcs.Auth, svcs.AccessToken)
	authHandler.RegisterRoutes(authGroup)

	convGroup := apiGroup.Group("/conversation", accessMidleware)
	convHandler := api.NewConversationHandler(svcs.Conversation)
	convHandler.RegisterRoutes(convGroup)

	userGroup := apiGroup.Group("/user", accessMidleware)
	userHandler := api.NewUserHandler(svcs.User)
	userHandler.RegisterRoutes(userGroup)

	return webServer
}
