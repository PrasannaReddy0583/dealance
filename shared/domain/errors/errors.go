package errors

import (
	"fmt"
	"net/http"
)

// AppError is the standard error type used across all Dealance services.
// Code and Message are returned to the client; Internal is logged but never exposed.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Internal   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// WithInternal attaches an internal error for logging purposes.
func (e *AppError) WithInternal(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Internal:   err,
	}
}

// --- 400 Bad Request ---

func ErrValidation(message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func ErrBadRequest(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// --- 401 Unauthorized ---

func ErrUnauthorized(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

var (
	ErrTokenExpired     = &AppError{Code: "TOKEN_EXPIRED", Message: "Token has expired", HTTPStatus: http.StatusUnauthorized}
	ErrTokenInvalid     = &AppError{Code: "TOKEN_INVALID", Message: "Token is invalid", HTTPStatus: http.StatusUnauthorized}
	ErrSignatureInvalid = &AppError{Code: "SIGNATURE_INVALID", Message: "Request signature is invalid", HTTPStatus: http.StatusUnauthorized}
	ErrDeviceNotTrusted = &AppError{Code: "DEVICE_NOT_TRUSTED", Message: "Device attestation failed", HTTPStatus: http.StatusUnauthorized}
	ErrReplayDetected   = &AppError{Code: "REPLAY_DETECTED", Message: "Replay attack detected", HTTPStatus: http.StatusUnauthorized}
	ErrOTPExpired       = &AppError{Code: "OTP_EXPIRED", Message: "OTP has expired", HTTPStatus: http.StatusUnauthorized}
	ErrOTPInvalid       = &AppError{Code: "OTP_INVALID", Message: "OTP is invalid", HTTPStatus: http.StatusUnauthorized}
	ErrRefreshRevoked   = &AppError{Code: "REFRESH_REVOKED", Message: "Refresh token has been revoked", HTTPStatus: http.StatusUnauthorized}
)

// --- 403 Forbidden ---

func ErrForbidden(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

var (
	ErrRoleForbidden         = &AppError{Code: "ROLE_FORBIDDEN", Message: "Insufficient role", HTTPStatus: http.StatusForbidden}
	ErrKYCRequired           = &AppError{Code: "KYC_REQUIRED", Message: "KYC verification required", HTTPStatus: http.StatusForbidden}
	ErrAccreditationRequired = &AppError{Code: "ACCREDITATION_REQUIRED", Message: "Investor accreditation required", HTTPStatus: http.StatusForbidden}
	ErrNDARequired           = &AppError{Code: "NDA_REQUIRED", Message: "NDA must be signed to access this resource", HTTPStatus: http.StatusForbidden}
)

// --- 404 Not Found ---

func ErrNotFound(resource string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// --- 409 Conflict ---

func ErrConflict(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// --- 429 Rate Limited ---

var ErrRateLimited = &AppError{
	Code:       "RATE_LIMITED",
	Message:    "Too many requests, please try again later",
	HTTPStatus: http.StatusTooManyRequests,
}

// --- 500 Internal ---

func ErrInternal() *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An unexpected error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}
}
