package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/response"
)

// AttestService defines the interface for device attestation verification.
type AttestService interface {
	VerifyiOSAttestation(ctx context.Context, token string) (bool, error)
	VerifyAndroidIntegrity(ctx context.Context, token string) (bool, error)
}

// AppAttest validates device attestation tokens (Apple App Attest / Google Play Integrity).
// In development mode with SKIP_ATTEST=true, this middleware is a no-op.
func AppAttest(rdb *redis.Client, attestSvc AttestService, log zerolog.Logger, skipAttest bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipAttest {
			c.Next()
			return
		}

		deviceID := c.GetHeader("X-Device-ID")
		if deviceID == "" {
			response.Error(c, apperrors.ErrDeviceNotTrusted)
			return
		}

		// Check iOS attestation
		iosToken := c.GetHeader("X-Attest-Token")
		androidToken := c.GetHeader("X-Integrity-Token")

		if iosToken == "" && androidToken == "" {
			response.Error(c, apperrors.ErrDeviceNotTrusted)
			return
		}

		ctx := c.Request.Context()

		// Determine which token to validate
		var token, platform string
		if iosToken != "" {
			token = iosToken
			platform = "ios"
		} else {
			token = androidToken
			platform = "android"
		}

		// Check Redis cache first
		tokenHash := sha256.Sum256([]byte(token))
		tokenHashPrefix := hex.EncodeToString(tokenHash[:])[:16]
		cacheKey := fmt.Sprintf("attest:%s:%s", deviceID, tokenHashPrefix)

		cached, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil && cached == "valid" {
			c.Next()
			return
		}

		// Cache miss — call vendor API
		var valid bool
		if platform == "ios" {
			valid, err = attestSvc.VerifyiOSAttestation(ctx, iosToken)
		} else {
			valid, err = attestSvc.VerifyAndroidIntegrity(ctx, androidToken)
		}

		if err != nil {
			log.Error().Err(err).Str("device_id", deviceID).Str("platform", platform).
				Msg("attestation verification failed")
			response.Error(c, apperrors.ErrDeviceNotTrusted)
			return
		}

		if !valid {
			response.Error(c, apperrors.ErrDeviceNotTrusted)
			return
		}

		// Cache successful attestation for 15 minutes
		rdb.Set(ctx, cacheKey, "valid", 15*time.Minute)

		c.Next()
	}
}
