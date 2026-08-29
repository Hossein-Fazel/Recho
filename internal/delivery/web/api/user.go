package api

import (
	"net/http"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService application.UserService
}

func NewUserHandler(user application.UserService) *UserHandler {
	return &UserHandler{
		userService: user,
	}
}

func (h *UserHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/search", h.Search)
}

// Search godoc
// @Summary Search in users by id
// @Description Returns similar usernames with some info
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string false "username"
// @Success 200 {object} dto.UserSearchResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/user/search [get]
func (h *UserHandler) Search(c echo.Context) error {
	query := c.QueryParam("q")

	if query == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"query is required",
		)
	}

	list, err := h.userService.Search(c.Request().Context(), query)
	if err != nil {
		return err
	}

	var res dto.UserSearchResponse

	for _, user := range list {
		res.Users = append(res.Users, uSearch2dtoUSearch(user))
	}

	return c.JSON(http.StatusOK, res)
}
