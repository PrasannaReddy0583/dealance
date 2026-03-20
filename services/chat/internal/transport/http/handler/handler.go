package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/chat/internal/application"
	"github.com/dealance/services/chat/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

type ChatHandler struct{ svc *application.ChatService }
func NewChatHandler(svc *application.ChatService) *ChatHandler { return &ChatHandler{svc: svc} }

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.CreateConversation(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetConversations(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.SendMessage(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	convID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	result, err := h.svc.GetMessages(c.Request.Context(), claims.Subject, convID, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *ChatHandler) EditMessage(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.EditMessage(c.Request.Context(), claims.Subject, c.Param("id"), req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Message edited"})
}

func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	if err := h.svc.DeleteMessage(c.Request.Context(), claims.Subject, c.Param("id")); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Message deleted"})
}

func (h *ChatHandler) MarkRead(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	if err := h.svc.MarkRead(c.Request.Context(), claims.Subject, c.Param("id")); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Marked as read"})
}

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-chat"})
}
