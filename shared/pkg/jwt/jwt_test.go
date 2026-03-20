package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	dealjwt "github.com/dealance/shared/pkg/jwt"
)

func writeTempPEM(t *testing.T, name string, pemType string, der []byte) string {
	t.Helper()
	f, err := os.CreateTemp("", name+"-*.pem")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	data := pem.EncodeToMemory(&pem.Block{Type: pemType, Bytes: der})
	if err := os.WriteFile(f.Name(), data, 0600); err != nil {
		t.Fatalf("write PEM: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func setupTestKeys(t *testing.T) (*dealjwt.Issuer, *dealjwt.Verifier) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	privPath := writeTempPEM(t, "priv", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))
	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal pub: %v", err)
	}
	pubPath := writeTempPEM(t, "pub", "PUBLIC KEY", pubDER)

	issuer, err := dealjwt.NewIssuer(privPath, "https://test.issuer.com", "https://test.audience.com")
	if err != nil {
		t.Fatalf("create issuer: %v", err)
	}

	verifier, err := dealjwt.NewVerifier(pubPath, "https://test.issuer.com", "https://test.audience.com")
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	return issuer, verifier
}

func TestIssueAndVerifyAccessToken(t *testing.T) {
	issuer, verifier := setupTestKeys(t)

	token, claims, err := issuer.IssueAccessToken(
		"user-123", "test@example.com",
		[]string{"ENTREPRENEUR"}, "ENTREPRENEUR", "APPROVED", "device-456", true,
	)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("token must not be empty")
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q, want user-123", claims.Subject)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email = %q", claims.Email)
	}
	if claims.ActiveRole != "ENTREPRENEUR" {
		t.Errorf("ActiveRole = %q", claims.ActiveRole)
	}

	verified, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if verified.Subject != "user-123" {
		t.Errorf("verified Subject = %q", verified.Subject)
	}
	if verified.KYCStatus != "APPROVED" {
		t.Errorf("KYCStatus = %q", verified.KYCStatus)
	}
	if !verified.EmailVerified {
		t.Error("EmailVerified should be true")
	}
}

func TestTokenExpirationInFuture(t *testing.T) {
	issuer, verifier := setupTestKeys(t)

	token, _, err := issuer.IssueAccessToken(
		"user-1", "a@b.com", []string{"INVESTOR"}, "INVESTOR", "PENDING", "d1", true,
	)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := verifier.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Unix() <= time.Now().Unix() {
		t.Error("expiration should be in the future")
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	_, verifier := setupTestKeys(t)

	_, err := verifier.Verify("not.a.token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestIssueRefreshToken(t *testing.T) {
	issuer, verifier := setupTestKeys(t)

	token, jti, err := issuer.IssueRefreshToken("user-123", "device-456", "family-789")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || jti == "" {
		t.Fatal("token and jti must not be empty")
	}

	claims, err := verifier.VerifyRefreshToken(token)
	if err != nil {
		t.Fatalf("VerifyRefreshToken: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q", claims.Subject)
	}
}

func TestGetPublicKey(t *testing.T) {
	issuer, _ := setupTestKeys(t)
	if issuer.GetPublicKey() == nil {
		t.Error("GetPublicKey returned nil")
	}
}

func TestWrongAudience(t *testing.T) {
	issuer, _ := setupTestKeys(t)

	// Create a verifier with a different audience
	pubKey := issuer.GetPublicKey()
	wrongVerifier := dealjwt.NewVerifierFromKey(pubKey, "https://test.issuer.com", "https://wrong.audience.com")

	token, _, _ := issuer.IssueAccessToken(
		"user-1", "a@b.com", []string{}, "", "", "", false,
	)

	_, err := wrongVerifier.Verify(token)
	if err == nil {
		t.Error("expected error for wrong audience")
	}
}
