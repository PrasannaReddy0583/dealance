package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionRepo implements SessionRepository using Redis.
type SessionRepo struct {
	rdb *redis.Client
}

func NewSessionRepo(rdb *redis.Client) *SessionRepo {
	return &SessionRepo{rdb: rdb}
}

// --- Signup Sessions ---

func (r *SessionRepo) CreateSignupSession(ctx context.Context, sessionID, email, otpHash string, ttl time.Duration) error {
	pipe := r.rdb.Pipeline()
	key := fmt.Sprintf("signup:session:%s", sessionID)
	pipe.HSet(ctx, key, "email", email, "otp_hash", otpHash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *SessionRepo) GetSignupSession(ctx context.Context, sessionID string) (string, string, error) {
	key := fmt.Sprintf("signup:session:%s", sessionID)
	result, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return "", "", err
	}
	if len(result) == 0 {
		return "", "", fmt.Errorf("session not found")
	}
	return result["email"], result["otp_hash"], nil
}

func (r *SessionRepo) DeleteSignupSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("signup:session:%s", sessionID)
	return r.rdb.Del(ctx, key).Err()
}

func (r *SessionRepo) UpdateSignupSessionOTP(ctx context.Context, sessionID, otpHash string) error {
	key := fmt.Sprintf("signup:session:%s", sessionID)
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("session not found")
	}
	return r.rdb.HSet(ctx, key, "otp_hash", otpHash).Err()
}

// --- Challenge Sessions (Passkey Login) ---

func (r *SessionRepo) CreateChallenge(ctx context.Context, challengeID string, challenge []byte, ttl time.Duration) error {
	key := fmt.Sprintf("challenge:%s", challengeID)
	return r.rdb.Set(ctx, key, challenge, ttl).Err()
}

func (r *SessionRepo) GetChallenge(ctx context.Context, challengeID string) ([]byte, error) {
	key := fmt.Sprintf("challenge:%s", challengeID)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *SessionRepo) DeleteChallenge(ctx context.Context, challengeID string) error {
	key := fmt.Sprintf("challenge:%s", challengeID)
	return r.rdb.Del(ctx, key).Err()
}

// --- Login OTP ---

func (r *SessionRepo) StoreLoginOTP(ctx context.Context, email, otpHash string, ttl time.Duration) error {
	key := fmt.Sprintf("login:otp:%s", email)
	return r.rdb.Set(ctx, key, otpHash, ttl).Err()
}

func (r *SessionRepo) GetLoginOTP(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf("login:otp:%s", email)
	return r.rdb.Get(ctx, key).Result()
}

func (r *SessionRepo) DeleteLoginOTP(ctx context.Context, email string) error {
	key := fmt.Sprintf("login:otp:%s", email)
	return r.rdb.Del(ctx, key).Err()
}

// --- JTI Blacklist ---

func (r *SessionRepo) BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("jti:blacklist:%s", jti)
	return r.rdb.Set(ctx, key, "1", ttl).Err()
}

func (r *SessionRepo) IsJTIBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("jti:blacklist:%s", jti)
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// --- JWKS Cache ---

func (r *SessionRepo) CacheJWKS(ctx context.Context, provider string, jwksData []byte, ttl time.Duration) error {
	key := fmt.Sprintf("jwks:%s", provider)
	return r.rdb.Set(ctx, key, jwksData, ttl).Err()
}

func (r *SessionRepo) GetCachedJWKS(ctx context.Context, provider string) ([]byte, error) {
	key := fmt.Sprintf("jwks:%s", provider)
	return r.rdb.Get(ctx, key).Bytes()
}
