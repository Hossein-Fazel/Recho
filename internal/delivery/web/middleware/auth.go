package middleware

import (
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/labstack/echo/v4"
)

const (
	CtxUsername  = "username"
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

func AccessMiddleware(accessToken application.AccessToken) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			cookie, err := c.Cookie(AccessToken)
			if err != nil || cookie.Value == "" {
				return echo.ErrUnauthorized
			}

			id, err := accessToken.Validate(cookie.Value)
			if err != nil {
				return echo.ErrUnauthorized
			}

			c.Set(CtxUsername, id)

			return next(c)
		}
	}
}

func RefreshMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			cookie, err := c.Cookie(RefreshToken)
			if err != nil || cookie.Value == "" {
				return echo.ErrUnauthorized
			}

			c.Set(RefreshToken, cookie.Value)

			return next(c)
		}
	}
}
