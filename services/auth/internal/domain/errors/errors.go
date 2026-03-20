package errors

import (
	apperrors "github.com/dealance/shared/domain/errors"
)

// Auth service specific errors

var (
	ErrEmailAlreadyExists    = apperrors.ErrConflict("Email is already registered")
	ErrUserNotFound          = apperrors.ErrNotFound("User")
	ErrStageInvalid          = apperrors.ErrBadRequest("STAGE_INVALID", "Cannot advance to this stage from current stage")
	ErrSessionNotFound       = apperrors.ErrBadRequest("SESSION_NOT_FOUND", "Signup session not found or expired")
	ErrSessionExpired        = apperrors.ErrBadRequest("SESSION_EXPIRED", "Signup session has expired")
	ErrCountryNotSupported   = apperrors.ErrBadRequest("COUNTRY_NOT_SUPPORTED", "This country is not supported")
	ErrSanctionsMatch        = apperrors.ErrForbidden("SANCTIONS_MATCH", "OFAC sanctions match detected")
	ErrKYCMaxAttempts        = apperrors.ErrForbidden("KYC_MAX_ATTEMPTS", "Maximum KYC verification attempts reached")
	ErrKYCAlreadyApproved    = apperrors.ErrConflict("KYC is already approved")
	ErrPasskeyInvalid        = apperrors.ErrUnauthorized("PASSKEY_INVALID", "Passkey verification failed")
	ErrPasskeyCloneDetected  = apperrors.ErrUnauthorized("PASSKEY_CLONE_DETECTED", "Possible cloned passkey detected")
	ErrOAuthTokenInvalid     = apperrors.ErrUnauthorized("OAUTH_TOKEN_INVALID", "OAuth token validation failed")
	ErrOAuthProviderError    = apperrors.ErrInternal()
	ErrProviderNotLinked     = apperrors.ErrNotFound("Identity provider")
	ErrProviderAlreadyLinked = apperrors.ErrConflict("Identity provider is already linked")
	ErrWebhookSignatureInvalid = apperrors.ErrUnauthorized("WEBHOOK_SIGNATURE_INVALID", "Webhook signature verification failed")
	ErrRefreshTokenNotFound  = apperrors.ErrUnauthorized("REFRESH_TOKEN_INVALID", "Refresh token not found")
	ErrChallengeExpired      = apperrors.ErrBadRequest("CHALLENGE_EXPIRED", "Authentication challenge has expired")
	ErrChallengeNotFound     = apperrors.ErrBadRequest("CHALLENGE_NOT_FOUND", "Authentication challenge not found")
)
