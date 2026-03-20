package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/deal/internal/application"
	"github.com/dealance/services/deal/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

type DealHandler struct{ svc *application.DealService }
func NewDealHandler(svc *application.DealService) *DealHandler { return &DealHandler{svc: svc} }

func (h *DealHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.CreateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.CreateDeal(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *DealHandler) Get(c *gin.Context) {
	result, err := h.svc.GetDeal(c.Request.Context(), c.Param("id"))
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) Update(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.UpdateDealRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.UpdateDeal(c.Request.Context(), c.Param("id"), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) GetByStartup(c *gin.Context) {
	result, err := h.svc.GetDealsByStartup(c.Request.Context(), c.Param("startup_id"))
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) GetMyDeals(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetMyDeals(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) Join(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.JoinDealRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.JoinDeal(c.Request.Context(), c.Param("id"), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Joined deal"})
}

func (h *DealHandler) Commit(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.CommitRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.CommitToDeal(c.Request.Context(), c.Param("id"), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Commitment recorded"})
}

func (h *DealHandler) SignNDA(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.SignNDARequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	ip := c.ClientIP(); ua := c.GetHeader("User-Agent")
	if err := h.svc.SignNDA(c.Request.Context(), c.Param("id"), claims.Subject, req, ip, ua); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "NDA signed"})
}

func (h *DealHandler) GetParticipants(c *gin.Context) {
	result, err := h.svc.GetParticipants(c.Request.Context(), c.Param("id"))
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) SendMessage(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.NegotiationMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.SendNegotiationMessage(c.Request.Context(), c.Param("id"), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Message sent"})
}

func (h *DealHandler) GetNegotiations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	result, err := h.svc.GetNegotiations(c.Request.Context(), c.Param("id"), limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) UploadDocument(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.UploadDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.UploadDocument(c.Request.Context(), c.Param("id"), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Document uploaded"})
}

func (h *DealHandler) GetDocuments(c *gin.Context) {
	result, err := h.svc.GetDocuments(c.Request.Context(), c.Param("id"))
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *DealHandler) GetMilestones(c *gin.Context) {
	result, err := h.svc.GetMilestones(c.Request.Context(), c.Param("id"))
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-deal"})
}
