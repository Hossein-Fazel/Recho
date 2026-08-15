package middleware

import (
	"github.com/Hossein-Fazel/Recho/internal/usecase"
	"github.com/labstack/echo/v4"
)

const (
	CtxUsername = "username"
	CookieToken = "access_token"
)

func AccessMiddleware(accessToken usecase.AccessToken) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			cookie, err := c.Cookie(CookieToken)
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
