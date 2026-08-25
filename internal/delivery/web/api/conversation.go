package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/application"
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
// @Param limit query int false "Number of conversations to return (default: 10)"
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
	cursor, err := decodeCursor(cursorParam)
	if cursorParam == "" && err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	conversations, err := h.convSvc.GetUserChats(
		c.Request().Context(),
		application.GetUserChatsParams{
			UserID:          userID,
			CursorID:        cursor.ID,
			CursorUpdatedAt: cursor.UpdatedAt,
			Limit:           limit,
		},
	)
	if err != nil {
		return err
	}

	last := conversations[len(conversations)-1]

	newCursor, err := encodeCursor(
		last.UpdatedAt,
		last.ConversationID,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, convList2convListRes(conversations, newCursor))
}
