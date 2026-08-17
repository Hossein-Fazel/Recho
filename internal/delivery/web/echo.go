package web

import (
	"fmt"

	"github.com/Hossein-Fazel/Recho/internal/delivery/web/api"
	"github.com/Hossein-Fazel/Recho/internal/usecase"

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

// @title Recho
// @version 0.0.0
// @description Realtime chat application
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func Start(authService usecase.AuthService, access usecase.AccessToken, conf Config) {
	pkg.Logger.Info().Msg("Initializing server")

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	e.Debug = conf.isDebug()

	e.Use(middleware.RequestLoggerWithConfig(
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
					Send()

				return nil
			},
		},
	))

	if e.Debug {
		swaggerRoute := fmt.Sprintf("/%s/*", conf.SwaggerRoute)
		e.GET(swaggerRoute, echoSwagger.WrapHandler)
	}

	apiGroup := e.Group("/api")
	api.Register(apiGroup, authService, access)

	pkg.Logger.Info().Msg("Starting server")

	e.Logger.Fatal(e.Start(":" + conf.Port))
}
