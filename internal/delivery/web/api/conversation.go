package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/labstack/echo/v4"
)

type ConversationHandler struct {
	convSvc application.ConverasionService
}

func NewConversationHandler(convSvc application.ConverasionService) *ConversationHandler {
	return &ConversationHandler{
		convSvc: convSvc,
	}
}

func (h *ConversationHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/", h.GetUserConversations)
}

// GetUserConversations godoc
// @Summary Get user conversations
// @Description Returns a paginated list of conversations belonging to the current user, ordered by most recent activity
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int true "Number of conversations to return (default: 10)"
// @Param cursor query string false "Pagination cursor for retrieving the next page. Use the cursor from the previous response to get the next page."
// @Success 200 {object} dto.UserConversationsResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversations/ [get]
func (h *ConversationHandler) GetUserConversations(c echo.Context) error {
	userID := c.Get(CtxUserID).(uuid.UUID)

	limit := int32(10)
	if value := c.QueryParam("limit"); value != "" {
		v, err := strconv.Atoi(value)
		if err != nil || v <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid limit")
		}

		limit = int32(v)
	}

	cursorParam := c.QueryParam("cursor")
	if cursorParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	conversations, newCursor, err := h.convSvc.GetUserConversations(
		c.Request().Context(),
		userID,
		cursorParam,
		limit,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, convList2convListRes(conversations, newCursor))
}

func (h *ConversationHandler) CreateDirectConversation(c echo.Context) error {
	var req dto.CreateDirectConversationRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "invalid request body",
		})
	}

	if req.UserID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, dto.ErrResponse{
			Error: "user_id is required",
		})
	}

	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, dto.ErrResponse{
			Error: "please login first",
		})
	}

	conversationID, err := h.convSvc.GetOrCreateDC(
		c.Request().Context(),
		userID,
		req.UserID,
	)
	if err != nil {
		return err
	}

	return c.JSON(
		http.StatusOK,
		dto.CreateDirectConversationResponse{
			ConversationID: conversationID,
		},
	)
}
