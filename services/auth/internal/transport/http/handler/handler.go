package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/dealance/services/auth/internal/application"
	"github.com/dealance/services/auth/internal/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/middleware"
	"github.com/dealance/shared/pkg/response"
)

var validate = validator.New()

// SignupHandler handles signup endpoints.
type SignupHandler struct {
	svc *application.SignupService
}

func NewSignupHandler(svc *application.SignupService) *SignupHandler {
	return &SignupHandler{svc: svc}
}

func (h *SignupHandler) Initiate(c *gin.Context) {
	var req entity.InitiateSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.Initiate(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SignupHandler) VerifyEmail(c *gin.Context) {
	var req entity.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SignupHandler) ResendOTP(c *gin.Context) {
	var req entity.ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.ResendOTP(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "OTP resent"})
}

func (h *SignupHandler) ConfirmAuth(c *gin.Context) {
	var req entity.ConfirmAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.ConfirmAuth(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Authentication method confirmed"})
}

func (h *SignupHandler) SetCountry(c *gin.Context) {
	var req entity.SetCountryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.SetCountry(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Country set"})
}

func (h *SignupHandler) SetRole(c *gin.Context) {
	var req entity.SetRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.SetRole(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Roles assigned"})
}

func (h *SignupHandler) InitiateKYC(c *gin.Context) {
	var req entity.InitiateKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := validate.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.InitiateKYC(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *SignupHandler) Status(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		// If JWT is available, use that
		claims := middleware.GetClaims(c)
		if claims != nil {
			userID = claims.Subject
		}
	}
	if userID == "" {
		response.Error(c, apperrors.ErrValidation("user_id query parameter required"))
		return
	}
	result, err := h.svc.GetSignupStatus(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// --- Login Handler ---

type LoginHandler struct {
	svc *application.LoginService
}

func NewLoginHandler(svc *application.LoginService) *LoginHandler {
	return &LoginHandler{svc: svc}
}

func (h *LoginHandler) BeginPasskeyLogin(c *gin.Context) {
	var req entity.BeginPasskeyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.BeginPasskeyLogin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *LoginHandler) FinishPasskeyLogin(c *gin.Context) {
	var req entity.FinishPasskeyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.FinishPasskeyLogin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *LoginHandler) OAuthLogin(c *gin.Context) {
	var req entity.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.OAuthLogin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *LoginHandler) BeginEmailLogin(c *gin.Context) {
	var req entity.BeginEmailLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	// Always return 200 for anti-enumeration
	_ = h.svc.BeginEmailLogin(c.Request.Context(), req)
	response.OK(c, map[string]string{"message": "If the email is registered, an OTP has been sent"})
}

func (h *LoginHandler) FinishEmailLogin(c *gin.Context) {
	var req entity.FinishEmailLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.FinishEmailLogin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// --- Token Handler ---

type TokenHandler struct {
	svc *application.TokenService
}

func NewTokenHandler(svc *application.TokenService) *TokenHandler {
	return &TokenHandler{svc: svc}
}

func (h *TokenHandler) Refresh(c *gin.Context) {
	var req entity.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *TokenHandler) Logout(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}

	var req entity.LogoutRequest
	_ = c.ShouldBindJSON(&req) // Optional body

	userID, _ := uuid.Parse(claims.Subject)
	if err := h.svc.Logout(c.Request.Context(), userID, claims.ID, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Logged out successfully"})
}

// --- Passkey Handler ---

type PasskeyHandler struct {
	svc *application.PasskeyService
}

func NewPasskeyHandler(svc *application.PasskeyService) *PasskeyHandler {
	return &PasskeyHandler{svc: svc}
}

func (h *PasskeyHandler) BeginRegistration(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}

	var req entity.BeginPasskeyRegistrationRequest
	_ = c.ShouldBindJSON(&req)

	result, err := h.svc.BeginRegistration(c.Request.Context(), claims.Subject, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *PasskeyHandler) FinishRegistration(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
		return
	}

	var req entity.FinishPasskeyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.FinishRegistration(c.Request.Context(), claims.Subject, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"message": "Passkey registered successfully"})
}

// --- KYC Webhook Handler ---

type KYCWebhookHandler struct {
	svc               *application.KYCWebhookService
	hypervergeSecret  string
	onfidoWebhookKey  string
}

func NewKYCWebhookHandler(svc *application.KYCWebhookService, hvSecret, onfidoKey string) *KYCWebhookHandler {
	return &KYCWebhookHandler{
		svc:              svc,
		hypervergeSecret: hvSecret,
		onfidoWebhookKey: onfidoKey,
	}
}

func (h *KYCWebhookHandler) Hyperverge(c *gin.Context) {
	// In production: verify HMAC signature BEFORE processing
	// signature := c.GetHeader("X-Hyperverge-Signature")
	// if !VerifyWebhookHMAC(body, signature, h.hypervergeSecret) { ... }

	var payload entity.HypervergeWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.ProcessHypervergeWebhook(c.Request.Context(), payload); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"status": "processed"})
}

func (h *KYCWebhookHandler) Onfido(c *gin.Context) {
	var payload entity.OnfidoWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.ProcessOnfidoWebhook(c.Request.Context(), payload); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, map[string]string{"status": "processed"})
}

// --- Health Handler ---

func HealthCheck(c *gin.Context) {
	response.OK(c, map[string]string{"status": "healthy", "service": "dealance-auth"})
}

