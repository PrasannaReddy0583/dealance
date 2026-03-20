package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateOTP generates a cryptographically secure 6-digit OTP.
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashOTP creates a SHA-256 hash of the OTP for storage.
// OTPs are never stored in plain text.
func HashOTP(otp string) string {
	h := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(h[:])
}

// VerifyOTP checks if the provided OTP matches the stored hash.
func VerifyOTP(otp, hash string) bool {
	return HashOTP(otp) == hash
}

// HashSHA256 returns the hex-encoded SHA-256 hash of the provided data.
func HashSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// HMACSHA256 computes an HMAC-SHA256 signature.
func HMACSHA256(key []byte, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMACSHA256 verifies an HMAC-SHA256 signature.
func VerifyHMACSHA256(key []byte, data []byte, signature string) bool {
	expected := HMACSHA256(key, data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// GenerateRandomBytes generates random bytes of the given length.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// GenerateRandomHex generates a random hex string of the given byte length.
func GenerateRandomHex(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateNonce generates a 16-byte random hex nonce.
func GenerateNonce() (string, error) {
	return GenerateRandomHex(16)
}
