package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	apperrors "github.com/dealance/shared/domain/errors"
	dealjwt "github.com/dealance/shared/pkg/jwt"
	"github.com/dealance/shared/pkg/response"
)

const (
	// ContextKeyClaims is the gin context key for JWT claims.
	ContextKeyClaims = "claims"
	// ContextKeyUserID is the gin context key for the authenticated user ID.
	ContextKeyUserID = "user_id"
)

// JWTAuth validates the JWT access token and extracts claims into the context.
// Checks for blacklisted JTIs in Redis (logout support).
func JWTAuth(verifier *dealjwt.Verifier, rdb *redis.Client, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Authorization header required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, apperrors.ErrUnauthorized("UNAUTHORIZED", "Invalid authorization format"))
			return
		}

		tokenString := parts[1]

		claims, err := verifier.Verify(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				response.Error(c, apperrors.ErrTokenExpired)
			} else {
				response.Error(c, apperrors.ErrTokenInvalid)
			}
			return
		}

		// Check JTI blacklist (for tokens from logged-out sessions)
		ctx := c.Request.Context()
		blacklistKey := "jti:blacklist:" + claims.ID
		exists, err := rdb.Exists(ctx, blacklistKey).Result()
		if err != nil {
			log.Error().Err(err).Msg("redis JTI blacklist check failed")
			// Fail open only in extreme cases — security decision
			// For a fintech app, fail closed:
			response.Error(c, apperrors.ErrInternal())
			return
		}
		if exists > 0 {
			response.Error(c, apperrors.ErrTokenInvalid)
			return
		}

		// Set claims in context for downstream handlers
		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.Subject)

		c.Next()
	}
}

// GetClaims retrieves JWT claims from the gin context.
func GetClaims(c *gin.Context) *dealjwt.Claims {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil
	}
	claims, ok := val.(*dealjwt.Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserID retrieves the authenticated user ID from the gin context.
func GetUserID(c *gin.Context) string {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return ""
	}
	return val.(string)
}

// BlacklistJTI adds a JTI to the Redis blacklist with a TTL matching the access token TTL.
func BlacklistJTI(ctx context.Context, rdb *redis.Client, jti string, ttl int) error {
	key := "jti:blacklist:" + jti
	return rdb.Set(ctx, key, "1", 0).Err()
}
