package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/labstack/echo/v4"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// Only apply our API error handling to /api/*.
	if strings.HasPrefix(c.Request().URL.Path, "/api/") {
		handleAPIError(err, c)
	}
}

func handleAPIError(err error, c echo.Context) {

	if appErr, ok := errors.AsType[*apperr.AppError](err); ok {
		c.JSON(appErr.Status, dto.ErrResponse{
			Error:  appErr.Message,
			Module: appErr.Module,
			Type:   string(appErr.Type),
		})

		return
	}

	if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		c.JSON(echoErr.Code, dto.ErrResponse{
			Type:  "echo error",
			Error: echoErrorMessage(echoErr),
		})

		return
	}

	c.JSON(http.StatusInternalServerError, dto.ErrResponse{
		Type:  string(apperr.ErrInternal),
		Error: "internal server error",
	})
}

func echoErrorMessage(err *echo.HTTPError) string {
	if msg, ok := err.Message.(string); ok {
		return msg
	}

	return http.StatusText(err.Code)
}
