package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/dealance/services/notify/internal/application"
	"github.com/dealance/services/notify/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()
type NotifyHandler struct{ svc *application.NotifyService }
func NewNotifyHandler(svc *application.NotifyService) *NotifyHandler { return &NotifyHandler{svc: svc} }

func (h *NotifyHandler) Send(c *gin.Context) {
	var req entity.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.Send(c.Request.Context(), req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Notification sent"})
}
func (h *NotifyHandler) GetNotifications(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetNotifications(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}
func (h *NotifyHandler) GetUnreadCount(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	count, err := h.svc.GetUnreadCount(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, map[string]int{"unread_count": count})
}
func (h *NotifyHandler) MarkRead(c *gin.Context) {
	if err := h.svc.MarkRead(c.Request.Context(), c.Param("id")); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Marked as read"})
}
func (h *NotifyHandler) MarkAllRead(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	if err := h.svc.MarkAllRead(c.Request.Context(), claims.Subject); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "All marked as read"})
}
func (h *NotifyHandler) RegisterDevice(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.RegisterDevice(c.Request.Context(), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Device registered"})
}
func (h *NotifyHandler) GetPreferences(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetPreferences(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}
func (h *NotifyHandler) UpdatePreferences(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.UpdatePreferences(c.Request.Context(), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Preferences updated"})
}
func HealthCheck(c *gin.Context) { response.OK(c, map[string]string{"status": "healthy", "service": "dealance-notify"}) }
