package web

import (
	"fmt"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/api"

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

// @title Recho
// @version 0.0.0
// @description Realtime chat application
// @BasePath /
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func Init(authService application.AuthService, access application.AccessToken, conf Config) *echo.Echo {
	pkg.Logger.Info().Msg("Initializing server")

	webServer := echo.New()
	webServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
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
	api.Register(apiGroup, authService, access)

	return webServer
}
