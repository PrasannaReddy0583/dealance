package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/dealance/services/wallet/internal/application"
	"github.com/dealance/services/wallet/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

type WalletHandler struct{ svc *application.WalletService }
func NewWalletHandler(svc *application.WalletService) *WalletHandler { return &WalletHandler{svc: svc} }

func (h *WalletHandler) GetWallet(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetOrCreateWallet(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetBalance(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.Deposit(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	result, err := h.svc.Withdraw(c.Request.Context(), claims.Subject, req)
	if err != nil { response.Error(c, err); return }
	response.Created(c, result)
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.Transfer(c.Request.Context(), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Transfer completed"})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetTransactions(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *WalletHandler) GetLedger(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	result, err := h.svc.GetLedger(c.Request.Context(), claims.Subject, limit)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *WalletHandler) AddBankAccount(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	var req entity.AddBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := validate.Struct(req); err != nil { response.Error(c, apperrors.ErrValidation(err.Error())); return }
	if err := h.svc.AddBankAccount(c.Request.Context(), claims.Subject, req); err != nil { response.Error(c, err); return }
	response.Created(c, map[string]string{"message": "Bank account added"})
}

func (h *WalletHandler) GetBankAccounts(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	result, err := h.svc.GetBankAccounts(c.Request.Context(), claims.Subject)
	if err != nil { response.Error(c, err); return }
	response.OK(c, result)
}

func (h *WalletHandler) RemoveBankAccount(c *gin.Context) {
	claims := middleware.GetClaims(c); if claims == nil { response.Error(c, apperrors.ErrValidation("auth required")); return }
	if err := h.svc.RemoveBankAccount(c.Request.Context(), claims.Subject, c.Param("id")); err != nil { response.Error(c, err); return }
	response.OK(c, map[string]string{"message": "Bank account removed"})
}

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-wallet"})
}
