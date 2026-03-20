package crypto_test

import (
	"testing"

	"github.com/dealance/shared/pkg/crypto"
)

func TestGenerateOTP(t *testing.T) {
	otp, err := crypto.GenerateOTP()
	if err != nil {
		t.Fatalf("GenerateOTP failed: %v", err)
	}
	if len(otp) != 6 {
		t.Errorf("OTP length = %d, want 6", len(otp))
	}
	// Verify it's numeric
	for _, c := range otp {
		if c < '0' || c > '9' {
			t.Errorf("OTP contains non-numeric character: %c", c)
		}
	}
}

func TestGenerateOTP_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := crypto.GenerateOTP()
		if err != nil {
			t.Fatalf("GenerateOTP failed: %v", err)
		}
		seen[otp] = true
	}
	// With 6-digit OTPs, 100 samples should have high uniqueness
	if len(seen) < 80 {
		t.Errorf("Only %d unique OTPs out of 100, expected more", len(seen))
	}
}

func TestHashOTP(t *testing.T) {
	otp := "123456"
	hash := crypto.HashOTP(otp)
	if hash == "" {
		t.Error("HashOTP returned empty string")
	}
	if hash == otp {
		t.Error("HashOTP returned the OTP itself (not hashed)")
	}
	// Deterministic
	hash2 := crypto.HashOTP(otp)
	if hash != hash2 {
		t.Error("HashOTP is not deterministic")
	}
}

func TestVerifyOTP(t *testing.T) {
	otp := "654321"
	hash := crypto.HashOTP(otp)

	if !crypto.VerifyOTP(otp, hash) {
		t.Error("VerifyOTP returned false for correct OTP")
	}
	if crypto.VerifyOTP("000000", hash) {
		t.Error("VerifyOTP returned true for incorrect OTP")
	}
}

func TestHMACSHA256(t *testing.T) {
	key := []byte("test-secret-key")
	data := []byte("test-data")

	sig := crypto.HMACSHA256(key, data)
	if sig == "" {
		t.Error("HMACSHA256 returned empty string")
	}

	// Verify
	if !crypto.VerifyHMACSHA256(key, data, sig) {
		t.Error("VerifyHMACSHA256 returned false for valid signature")
	}

	// Wrong key
	if crypto.VerifyHMACSHA256([]byte("wrong-key"), data, sig) {
		t.Error("VerifyHMACSHA256 returned true for wrong key")
	}

	// Wrong data
	if crypto.VerifyHMACSHA256(key, []byte("wrong-data"), sig) {
		t.Error("VerifyHMACSHA256 returned true for wrong data")
	}
}

func TestHashSHA256(t *testing.T) {
	data := []byte("hello world")
	hash := crypto.HashSHA256(data)
	if hash == "" {
		t.Error("HashSHA256 returned empty string")
	}
	// Known SHA-256 of "hello world"
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("HashSHA256 = %s, want %s", hash, expected)
	}
}

func TestGenerateRandomHex(t *testing.T) {
	hex, err := crypto.GenerateRandomHex(16)
	if err != nil {
		t.Fatalf("GenerateRandomHex failed: %v", err)
	}
	if len(hex) != 32 {
		t.Errorf("hex length = %d, want 32 (16 bytes = 32 hex chars)", len(hex))
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}
	if len(nonce) != 32 {
		t.Errorf("nonce length = %d, want 32", len(nonce))
	}
	// Unique
	nonce2, _ := crypto.GenerateNonce()
	if nonce == nonce2 {
		t.Error("Two consecutive nonces should be different")
	}
}
