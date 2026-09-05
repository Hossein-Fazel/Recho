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
convSvc *application.ConversationService
}

func NewConversationHandler(convSvc *application.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		convSvc: convSvc,
	}
}

func (h *ConversationHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/", h.GetUserConversations)
	g.POST("/", h.GetorCreateDirectConversation)
	g.GET("/:id", h.GetConversationByID)
	g.GET("/:id/messages", h.GetConversationMessages)
	g.GET("/:id/info", h.GetConversationInfo)
	g.GET("/:id/members", h.GetGroupMembers)
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

// GetConversationByID godoc
// @Summary Get a single conversation's summary
// @Description Returns info about one conversation the caller belongs to: for a direct chat this includes the other user's username, display name and avatar; for a group it includes the group name and avatar. Useful when a client receives a message for a conversation it doesn't have cached yet (e.g. a first message from a new person) and needs to render it in the chat list.
// @Tags conversation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} dto.Conversation
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversation/{id} [get]
func (h *ConversationHandler) GetConversationByID(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid conversation id", err)
	}

	conversation, err := h.convSvc.GetConversationByID(
		c.Request().Context(),
		userID,
		conversationID,
	)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, conv2convRes(conversation))
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

// GetConversationInfo godoc
// @Summary Get a single conversation's info
// @Tags conversation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} dto.Conversation
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 404 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversation/{id}/info [get]
func (h *ConversationHandler) GetConversationInfo(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid conversation id", err)
	}

	info, err :=
		h.convSvc.GetConversationInfo(
			c.Request().Context(),
			userID,
			conversationID,
		)

	if err != nil {
		return err
	}

	response := dto.ConversationInfo{
		ConversationID:   info.ID,
		ConversationType: info.ConversationType,
	}

	if info.User != nil {
		response.User = &dto.UserInfo{
			ID:          info.User.UserID,
			Username:    info.User.Username,
			DisplayName: info.User.DisplayName,
			AvatarURL:   info.User.AvatarUrl,
			Bio:         info.User.Bio,
		}
	}
	if info.Group != nil {
		group := &dto.GroupInfo{
			ID:          info.Group.GroupID,
			Name:        info.Group.GroupName,
			AvatarURL:   info.Group.GroupAvatarUrl,
			Bio:         info.Group.GroupBio,
			Role:        info.Group.ViewerRole,
			MemberCount: info.Group.MemberCount,
		}

		if info.Group.CanShareInvite() {
			group.InviteCode = info.Group.InviteCode
		}

		response.Group = group
	}

	return c.JSON(http.StatusOK, response)
}


// GetGroupMembers godoc
// @Summary Get group members
// @Description Returns a paginated list of group members, ordered by userIDs
// @Tags conversation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int true "Number of members to return (default: 10)"
// @Param cursor query string false "Pagination cursor for retrieving the next page. Use the cursor from the previous response to get the next page."
// @Success 200 {object} dto.UserConversationsResponse
// @Failure 400 {object} dto.ErrResponse
// @Failure 401 {object} dto.ErrResponse
// @Failure 500 {object} dto.ErrResponse
// @Router /api/conversations/{id}/members [get]
func (h *ConversationHandler) GetGroupMembers(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperr.InvalidInput("web", "invalid conversation id", err)
	}

	limit := 10

	if value := c.QueryParam("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return apperr.InvalidInput("web", "invalid limit", err)
		}
	}

	if limit > 50 {
		limit = 50
	}

	cursor := c.QueryParam("cursor")

	members, nextCursor, err :=
		h.convSvc.GetGroupMembers(
			c.Request().Context(),
			application.GetGroupMembersParams{
				GroupID: conversationID,
				UserID:  userID,
				Cursor:  cursor,
				Limit:   int32(limit),
			},
		)

	if err != nil {
		return err
	}

	response := dto.GetGroupMembersResponse{
		Members:    make([]*dto.GroupMember, 0, len(members)),
		NextCursor: nextCursor,
	}

	for _, member := range members {
		response.Members = append(
			response.Members,
			&dto.GroupMember{
				UserID:      member.UserID,
				Username:    member.Username,
				DisplayName: member.DisplayName,
				AvatarURL:   member.AvatarURL,
				Role:        member.Role,
			},
		)
	}

	return c.JSON(http.StatusOK, response)
}
