package handler

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/dealance/services/media/internal/application"
	"github.com/dealance/services/media/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()
type MediaHandler struct{ svc *application.MediaService }
func NewMediaHandler(svc *application.MediaService) *MediaHandler { return &MediaHandler{svc: svc} }

func (h *MediaHandler) RequestUploadURL(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.RequestUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.RequestUploadURL(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}
func (h *MediaHandler) ConfirmUpload(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.ConfirmUpload(c.Request.Context(), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Upload confirmed"})
}
func (h *MediaHandler) GetMyUploads(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetMyUploads(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}
func (h *MediaHandler) DeleteUpload(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	if err := h.svc.DeleteUpload(c.Request.Context(), claims.Subject, c.Param("id")); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Upload deleted"})
}
func HealthCheck(c *gin.Context) { response.OK(c, map[string]string{"status": "healthy", "service": "dealance-media"}) }
