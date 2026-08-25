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
	g.GET("/", h.GetConversations)
}

func (h *ConversationHandler) GetConversations(c echo.Context) error {
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
