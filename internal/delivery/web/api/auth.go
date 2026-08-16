package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"

	"github.com/Hossein-Fazel/Recho/internal/delivery/web/middleware"
	"github.com/Hossein-Fazel/Recho/internal/usecase"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService usecase.AuthService
	accessToken usecase.AccessToken
}

func NewAuthHandler(auth usecase.AuthService, accessToken usecase.AccessToken) *AuthHandler {
	return &AuthHandler{
		authService: auth,
		accessToken: accessToken,
	}
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/login", h.Login)
	g.POST("/register", h.Register)
	g.GET("/refresh", h.Refresh, middleware.RefreshMiddleware())
	g.GET("/verify", h.Verify, middleware.AccessMiddleware(h.accessToken))
	g.POST("/logout", h.Logout, middleware.AccessMiddleware(h.accessToken))
}

// Register godoc
// @Summary Register a new account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegitsterRequest true "User ID"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "invalid request body",
		})
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "username and password are required",
		})
	}

	result, err := h.authService.Register(
		c.Request().Context(),
		req.Username,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUsernameExists):
			return c.JSON(http.StatusConflict, dto.ErrResponse{
				Error: "username already exists",
			})

		default:
			return c.JSON(http.StatusInternalServerError, dto.ErrResponse{
				Error: "internal server error",
			})
		}
	}

	// Set access token cookie
	c.SetCookie(h.generateCookie(
		"access_token",
		result.AccessToken.Token,
		result.AccessToken.TTL,
	))

	// Set refresh token cookie
	c.SetCookie(h.generateCookie(
		"refresh_token",
		result.RefreshToken.Token,
		result.RefreshToken.TTL,
	))

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Message: "register successfully",
		User:    result.User,
	})
}

// Login godoc
// @Summary Login to account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User ID"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "invalid request body",
		})
	}

	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "username and password is required",
		})
	}

	result, err := h.authService.Login(
		c.Request().Context(),
		req.Username,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidLogin):
			return c.JSON(http.StatusUnauthorized, dto.ErrResponse{
				Error: "invalid username or password",
			})

		default:
			return c.JSON(http.StatusInternalServerError, dto.ErrResponse{
				Error: "internal server error",
			})
		}
	}

	// Set access token cookie
	c.SetCookie(h.generateCookie(
		"access_token",
		result.AccessToken.Token,
		result.AccessToken.TTL,
	))

	// Set refresh token cookie
	c.SetCookie(h.generateCookie(
		"refresh_token",
		result.RefreshToken.Token,
		result.RefreshToken.TTL,
	))

	return c.JSON(http.StatusOK, dto.AuthResponse{
		Message: "login successfully",
		User:    result.User,
	})
}

// Refresh godoc
// @Summary Refrsh tokens
// @Tags auth
// @Produce json
// @Security CookieAuth
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrResponse
// @Router /api/auth/refresh [get]
func (h *AuthHandler) Refresh(c echo.Context) error {
	RefreshToken := strings.TrimSpace(c.Get("refresh_token").(string))

	at, rt, err := h.authService.Refresh(
		c.Request().Context(),
		RefreshToken,
	)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidLogin):
			return c.JSON(http.StatusUnauthorized, dto.ErrResponse{
				Error: "invalid refresh token",
			})

		default:
			return c.JSON(http.StatusInternalServerError, dto.ErrResponse{
				Error: "internal server error",
			})
		}
	}

	c.SetCookie(h.deleteCookie("access_token"))
	c.SetCookie(h.deleteCookie("refresh_token"))

	c.SetCookie(h.generateCookie(
		"access_token",
		at.Token,
		at.TTL,
	))

	c.SetCookie(h.generateCookie(
		"refresh_token",
		rt.Token,
		rt.TTL,
	))

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "tokens refresh successfully",
	})
}

// Verify godoc
// @Summary Verify authentication
// @Description Check if user is authenticated
// @Tags auth
// @Produce json
// @Security CookieAuth
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrResponse
// @Router /api/auth/verify [get]
func (a *AuthHandler) Verify(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "authenticated",
	})
}

// Logout godoc
// @Summary Logout user
// @Description Clear authentication cookie
// @Tags auth
// @Produce json
// @Security CookieAuth
// @Success 200 {object} dto.MessageResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	c.SetCookie(h.deleteCookie("access_token"))
	c.SetCookie(h.deleteCookie("refresh_token"))

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "logged out successfully",
	})
}

func (h *AuthHandler) deleteCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (h *AuthHandler) generateCookie(name string, token string, maxAge time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
