package api

import (
	"net/http"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(user *application.UserService) *UserHandler {
	return &UserHandler{
		userService: user,
	}
}

func (h *UserHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/me", h.GetMe)
	g.PATCH("/me", h.UpdateProfile)
	g.POST("/me/avatar", h.UploadAvatar)
	g.GET("/search", h.Search)
	g.GET("/:id", h.GetUserInfo)
}

// GetMe godoc
// @Summary Get current user profile
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.User
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/user/me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	user, err := h.userService.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user2dtoUser(user))
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Description Updates the username, display name and/or bio. Omitted fields stay unchanged.
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "profile"
// @Success 200 {object} dto.User
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/user/me [patch]
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return apperr.InvalidInput("web", "invalid request body", err)
	}

	user, err := h.userService.UpdateProfile(
		c.Request().Context(),
		userID,
		application.UpdateProfileParams{
			Username:    req.Username,
			DisplayName: req.DisplayName,
			Bio:         req.Bio,
		},
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user2dtoUser(user))
}

// UploadAvatar godoc
// @Summary Update current user avatar
// @Description Replaces the current user's avatar with the uploaded image.
// @Tags user
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "avatar image"
// @Success 200 {object} dto.User
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/user/me/avatar [post]
func (h *UserHandler) UploadAvatar(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	header, err := c.FormFile("avatar")
	if err != nil {
		return apperr.InvalidInput("web", "avatar file is required", err)
	}

	file, err := header.Open()
	if err != nil {
		return apperr.Internal("web", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			c.Logger().Errorf("failed to close avatar upload: %v", err)
		}
	}()

	user, err := h.userService.UpdateAvatar(
		c.Request().Context(),
		userID,
		model.UploadedFile{
			Content:  file,
			Size:     header.Size,
			FileName: header.Filename,
		},
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user2dtoUser(user))
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

// GetUserInfo godoc
// @Summary get a user info by
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.User
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/user/{id} [get]
func (h *UserHandler) GetUserInfo(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid group id", err)
	}

	user, err := h.userService.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user2dtoUser(user))
}
