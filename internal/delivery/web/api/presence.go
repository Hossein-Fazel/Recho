package api

import (
	"net/http"
	"strings"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PresenceHandler struct {
	presenceService *application.PresenceService
}

func NewPresenceHandler(presenceService *application.PresenceService) *PresenceHandler {
	return &PresenceHandler{
		presenceService: presenceService,
	}
}

func (h *PresenceHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetPresence)
	g.POST("/subscription", h.Subscribe)
	g.DELETE("/subscription", h.Unsubscribe)
}

// GetPresence godoc
// @Summary Get batch users status
// @Tags presence
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.PresenceResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/presence [get]
func (h *PresenceHandler) GetPresence(c echo.Context) error {
	_, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	userIDsParam := c.QueryParam("user_ids")
	if userIDsParam == "" {
		return apperr.InvalidInput("presence handler", "user ids required", nil)
	}

	rawIDs := strings.Split(userIDsParam, ",")

	userIDs := make([]uuid.UUID, 0, len(rawIDs))

	for _, rawID := range rawIDs {
		userID, err := uuid.Parse(rawID)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"invalid user_id",
			)
		}

		userIDs = append(userIDs, userID)
	}

	statuses, err := h.presenceService.GetStatuses(userIDs)
	if err != nil {
		return err
	}

	response := make([]dto.Presence, 0, len(statuses))

	for _, status := range statuses {
		response = append(response, dto.Presence{
			UserID: status.UserID,
			Online: status.Online,
		})
	}

	return c.JSON(http.StatusOK, dto.PresenceResponse{Statuses: response})
}

// Subscribe godoc
// @Summary subscribe batch of users
// @Tags presence
// @Produce json
// @Security BearerAuth
// @Param request body dto.SubscribePresenceRequest true "Users ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/presence//subscriptions [post]
func (h *PresenceHandler) Subscribe(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	var req dto.SubscribePresenceRequest

	if err := c.Bind(&req); err != nil {
		return apperr.InvalidInput("presence handler", "user ids required", nil)
	}

	if len(req.UserIDs) == 0 {
		return apperr.InvalidInput("presence handler", "user_ids cannot be empty", nil)
	}

	if err := h.presenceService.Subscribe(
		userID,
		req.UserIDs,
	); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "subscribed successfully",
	})
}

// Unsubscribe godoc
// @Summary Unsubscribe batch of users
// @Tags presence
// @Produce json
// @Security BearerAuth
// @Param request body dto.UnsubscribePresenceRequest true "Users ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/presence//subscriptions [delete]
func (h *PresenceHandler) Unsubscribe(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	var req dto.UnsubscribePresenceRequest

	if err := c.Bind(&req); err != nil {
		return apperr.InvalidInput("presence handler", "user ids required", nil)
	}

	if len(req.UserIDs) == 0 {
		return apperr.InvalidInput("presence handler", "user_ids cannot be empty", nil)
	}

	if err := h.presenceService.Unsubscribe(
		userID,
		req.UserIDs,
	); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "unsubscribed successfully",
	})
}
