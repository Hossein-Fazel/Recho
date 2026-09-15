package api

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/labstack/echo/v4"
)

type GroupHandler struct {
	groupSvc *application.GroupService
}

func NewGroupHandler(groupSvc *application.GroupService) *GroupHandler {
	return &GroupHandler{
		groupSvc: groupSvc,
	}
}

func (h *GroupHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/", h.CreateGroup)
	g.GET("/invite/:code", h.GetGroupByInviteCode)
	g.POST("/join", h.JoinGroup)
	g.POST("/:id/leave", h.LeaveGroup)
	g.DELETE("/:id", h.DeleteGroup)
}

// CreateGroup godoc
// @Summary Create a group
// @Description Creates a group owned by the current user and assigns it an invite code
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGroupRequest true "group"
// @Success 201 {object} dto.CreateGroupResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/group/ [post]
func (h *GroupHandler) CreateGroup(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	var req dto.CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		return apperr.InvalidInput("web", "invalid request body", err)
	}

	group, err := h.groupSvc.Create(
		c.Request().Context(),
		userID,
		req.Name,
		req.Bio,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.CreateGroupResponse{
		ConversationID: group.ID,
		Name:           group.Name,
		Bio:            group.Bio,
		AvatarURL:      group.AvatarURL,
		InviteCode:     group.InviteCode,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
	})
}

// GetGroupByInviteCode godoc
// @Summary Preview a group from an invite code
// @Description Returns the public details of the group behind an invite code so an invited user can decide whether to join. Does not require membership.
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Invite code"
// @Success 200 {object} dto.GroupPreviewResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/group/invite/{code} [get]
func (h *GroupHandler) GetGroupByInviteCode(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	preview, err := h.groupSvc.PreviewByInviteCode(
		c.Request().Context(),
		userID,
		c.Param("code"),
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, groupPreview2res(preview))
}

// JoinGroup godoc
// @Summary Join a group using an invite code
// @Description Adds the current user to the group behind the invite code. Joining a group the user already belongs to succeeds without changing anything.
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JoinGroupRequest true "invite code"
// @Success 200 {object} dto.GroupPreviewResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/group/join [post]
func (h *GroupHandler) JoinGroup(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	var req dto.JoinGroupRequest
	if err := c.Bind(&req); err != nil {
		return apperr.InvalidInput("web", "invalid request body", err)
	}

	preview, err := h.groupSvc.JoinByInviteCode(
		c.Request().Context(),
		userID,
		req.InviteCode,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, groupPreview2res(preview))
}

// LeaveGroup godoc
// @Summary Leave a group
// @Description Removes the current user from the group.
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/group/{id}/leave [post]
func (h *GroupHandler) LeaveGroup(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid group id", err)
	}

	err = h.groupSvc.LeaveGroup(c.Request().Context(), userID, groupID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "You left the group successfully",
	})
}

// DeleteGroup godoc
// @Summary Delete a group
// @Description Delete the group.
// @Tags group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/group/{id} [delete]
func (h *GroupHandler) DeleteGroup(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid group id", err)
	}

	err = h.groupSvc.Delete(c.Request().Context(), userID, groupID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "the group was deleted successfully",
	})
}
