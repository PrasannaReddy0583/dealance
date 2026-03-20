package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/response"
)

// NDAChecker provides the ability to check NDA signed status for a deal room.
type NDAChecker interface {
	IsNDASigned(ctx context.Context, userID string, dealRoomID string) (bool, error)
}

// RequireAuth ensures the request has a valid JWT with claims in context.
// This is usually redundant if JWTAuth middleware is applied, but serves as
// an explicit assertion for route clarity.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
			return
		}
		c.Next()
	}
}

// RequireRole checks that the user has the specified role.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
			return
		}

		hasRole := false
		for _, r := range claims.Roles {
			if r == role || r == "ADMIN" || r == "BOTH" {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Error(c, apperrors.ErrRoleForbidden)
			return
		}

		c.Next()
	}
}

// RequireKYCApproved ensures the user has completed KYC verification.
func RequireKYCApproved() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
			return
		}

		if claims.KYCStatus != "APPROVED" {
			response.Error(c, apperrors.ErrKYCRequired)
			return
		}

		c.Next()
	}
}

// RequireInvestorVerified checks that the user has full investor accreditation.
// This is a tighter gate than RequireKYCApproved — it requires investor_verifications.status == VERIFIED.
func RequireInvestorVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
			return
		}

		// Check role first
		hasInvestorRole := false
		for _, r := range claims.Roles {
			if r == "INVESTOR" || r == "BOTH" {
				hasInvestorRole = true
				break
			}
		}
		if !hasInvestorRole {
			response.Error(c, apperrors.ErrRoleForbidden)
			return
		}

		// KYC must be approved
		if claims.KYCStatus != "APPROVED" {
			response.Error(c, apperrors.ErrAccreditationRequired)
			return
		}

		// Note: Full investor verification status check requires a DB call
		// which should be done in the handler/service layer, not middleware.
		// The JWT claim only carries basic KYC status. For full accreditation,
		// the service layer queries investor_verifications table.

		c.Next()
	}
}

// RequireNDASigned checks that the user has signed the NDA for a given deal room.
// The deal room ID is extracted from the URL parameter.
func RequireNDASigned(checker NDAChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authentication required"))
			return
		}

		roomID := c.Param("id")
		if roomID == "" {
			response.Error(c, apperrors.ErrValidation("Deal room ID required"))
			return
		}

		signed, err := checker.IsNDASigned(c.Request.Context(), claims.Subject, roomID)
		if err != nil {
			response.Error(c, apperrors.ErrInternal())
			return
		}

		if !signed {
			response.Error(c, apperrors.ErrNDARequired)
			return
		}

		c.Next()
	}
}
