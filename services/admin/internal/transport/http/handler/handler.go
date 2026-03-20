package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/dealance/services/admin/internal/application"
	"github.com/dealance/services/admin/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()
type AdminHandler struct{ svc *application.AdminService }
func NewAdminHandler(svc *application.AdminService) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetDashboardStats(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}
func (h *AdminHandler) ModerateContent(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.ContentModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.ModerateContent(c.Request.Context(), claims.Subject, req, c.ClientIP()); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Moderation action applied"})
}
func (h *AdminHandler) GetAuditLog(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	result, err := h.svc.GetAuditLog(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}
func HealthCheck(c *gin.Context) { response.OK(c, map[string]string{"status": "healthy", "service": "dealance-admin"}) }
