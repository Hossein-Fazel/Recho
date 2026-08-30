package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/web/dto"
	"github.com/labstack/echo/v4"
)

type ConversationHandler struct {
	convSvc *application.ConverasionService
}

func NewConversationHandler(convSvc *application.ConverasionService) *ConversationHandler {
	return &ConversationHandler{
		convSvc: convSvc,
	}
}

func (h *ConversationHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/", h.GetUserConversations)
	g.POST("/", h.GetorCreateDirectConversation)
	g.GET("/:id/messages", h.GetConversationMessages)
}

// GetUserConversations godoc
// @Summary Get user conversations
// @Description Returns a paginated list of conversations belonging to the current user, ordered by most recent activity
// @Tags conversation
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

// GetOrCreateDirectConversation godoc
// @Summary if a DC exists between two users return it's id else create a new one and then return it's id
// @Tags conversation
// @Accept json
// @Produce json
// @Param request body dto.GetOrCreateDirectConversationRequest true "params"
// @Success 200 {object} dto.GetOrCreateDirectConversationResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversation/ [post]
func (h *ConversationHandler) GetorCreateDirectConversation(c echo.Context) error {
	var req dto.GetOrCreateDirectConversationRequest

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
		dto.GetOrCreateDirectConversationResponse{
			ConversationID: conversationID,
		},
	)
}

// GetConversationMessages godoc
// @Summary Get conversation messages
// @Description Returns a paginated list of conversation message
// @Tags conversation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int true "Number of messages to return (default: 20)"
// @Param cursor query string false "Pagination cursor for retrieving the next page. Use the cursor from the previous response to get the next page."
// @Success 200 {object} dto.UserConversationsResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversations/:id/messages [get]
func (h *ConversationHandler) GetConversationMessages(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid conversation id", err)
	}

	limit := 20

	if value := c.QueryParam("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return apperr.InvalidInput("web", "invalid limit", err)
		}
	}

	if limit > 50 {
		limit = 50
	}

	cursor := c.QueryParam("cursor_id")

	messages, nextCursor, err :=
		h.convSvc.GetConversationMessages(
			c.Request().Context(),
			userID,
			conversationID,
			cursor,
			int32(limit),
		)

	if err != nil {
		return err
	}

	response := dto.GetConversationMessagesResponse{
		Messages:   make([]*dto.Message, 0, len(messages)),
		NextCursor: nextCursor,
	}

	for _, message := range messages {
		response.Messages = append(
			response.Messages,
			&dto.Message{
				ID:             message.ID,
				ConversationID: message.ConversationID,
				SenderID:       message.SenderID,
				Content:        message.Content,
				CreatedAt:      message.CreatedAt,
				UpdatedAt:      message.UpdatedAt,
			},
		)
	}

	return c.JSON(http.StatusOK, response)
}
