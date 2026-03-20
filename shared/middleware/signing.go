package middleware

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/crypto"
	"github.com/dealance/shared/pkg/response"
)

// DeviceKeyLookup provides device public keys for HMAC verification.
type DeviceKeyLookup interface {
	GetDeviceSigningKey(ctx context.Context, deviceID string) ([]byte, error)
}

// RequestSigning validates device-bound HMAC request signatures.
// Checks: timestamp drift (±30s), nonce replay (Redis), HMAC-SHA256 signature.
// In development mode with SKIP_SIGNING=true, this middleware is a no-op.
func RequestSigning(rdb *redis.Client, keyLookup DeviceKeyLookup, log zerolog.Logger, skipSigning bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipSigning {
			c.Next()
			return
		}

		deviceID := c.GetHeader("X-Device-ID")
		timestampStr := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signature := c.GetHeader("X-Signature")

		if deviceID == "" || timestampStr == "" || nonce == "" || signature == "" {
			response.Error(c, apperrors.ErrSignatureInvalid)
			return
		}

		ctx := c.Request.Context()

		// 1. Check timestamp drift (±30 seconds)
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.Error(c, apperrors.ErrSignatureInvalid)
			return
		}

		now := time.Now().Unix()
		if math.Abs(float64(now-timestamp)) > 30 {
			response.Error(c, apperrors.ErrSignatureInvalid)
			return
		}

		// 2. Check nonce replay
		nonceKey := fmt.Sprintf("nonce:%s", nonce)
		exists, err := rdb.Exists(ctx, nonceKey).Result()
		if err != nil {
			log.Error().Err(err).Msg("redis nonce check failed")
			response.Error(c, apperrors.ErrInternal())
			return
		}
		if exists > 0 {
			response.Error(c, apperrors.ErrReplayDetected)
			return
		}

		// 3. Store nonce with TTL 120s
		rdb.Set(ctx, nonceKey, "1", 120*time.Second)

		// 4. Load device signing key
		deviceKey, err := keyLookup.GetDeviceSigningKey(ctx, deviceID)
		if err != nil {
			log.Error().Err(err).Str("device_id", deviceID).Msg("device key lookup failed")
			response.Error(c, apperrors.ErrSignatureInvalid)
			return
		}

		// 5. Compute and verify HMAC
		// Read body for hashing (and put it back)
		var bodyHash string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				response.Error(c, apperrors.ErrSignatureInvalid)
				return
			}
			bodyHash = crypto.HashSHA256(bodyBytes)
			// Replace body so handlers can read it
			c.Request.Body = io.NopCloser(
				&bytesReader{data: bodyBytes, offset: 0},
			)
		} else {
			bodyHash = crypto.HashSHA256([]byte(""))
		}

		message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
			c.Request.Method,
			c.Request.URL.Path,
			bodyHash,
			timestampStr,
			nonce,
		)

		if !crypto.VerifyHMACSHA256(deviceKey, []byte(message), signature) {
			response.Error(c, apperrors.ErrSignatureInvalid)
			return
		}

		c.Next()
	}
}

// bytesReader is a simple reader that allows re-reading body bytes.
type bytesReader struct {
	data   []byte
	offset int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}
