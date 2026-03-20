package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/dealance/shared/domain/entity"
)

// Claims represents the standard JWT claims for Dealance.
// Standard fields (sub, iss, aud, exp, iat, jti) live in RegisteredClaims.
// Access them via: claims.Subject, claims.ID, claims.ExpiresAt, etc.
// Custom fields below are app-specific.
type Claims struct {
	Email         string   `json:"email"`
	Roles         []string `json:"roles"`
	ActiveRole    string   `json:"active_role"`
	KYCStatus     string   `json:"kyc_status"`
	EmailVerified bool     `json:"email_verified"`
	DeviceID      string   `json:"device_id"`
	jwt.RegisteredClaims
}

// TokenPair contains the access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// Issuer handles JWT token creation with the private key (auth service only).
type Issuer struct {
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
}

// Verifier handles JWT token validation with the public key (all services).
type Verifier struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

// NewIssuer creates a new JWT issuer from a PEM private key file.
func NewIssuer(privateKeyPath, issuer, audience string) (*Issuer, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key (PKCS1: %v, PKCS8: %v)", err, err2)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		privateKey = rsaKey
	}

	return &Issuer{
		privateKey: privateKey,
		issuer:     issuer,
		audience:   audience,
	}, nil
}

// NewVerifier creates a new JWT verifier from a PEM public key file.
func NewVerifier(publicKeyPath, issuer, audience string) (*Verifier, error) {
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return &Verifier{
		publicKey: publicKey,
		issuer:    issuer,
		audience:  audience,
	}, nil
}

// NewVerifierFromKey creates a verifier directly from an RSA public key.
func NewVerifierFromKey(publicKey *rsa.PublicKey, issuer, audience string) *Verifier {
	return &Verifier{
		publicKey: publicKey,
		issuer:    issuer,
		audience:  audience,
	}
}

// IssueAccessToken creates a signed access token with the given claims.
func (i *Issuer) IssueAccessToken(userID, email string, roles []string, activeRole, kycStatus, deviceID string, emailVerified bool) (string, *Claims, error) {
	now := time.Now()
	jti := uuid.New().String()
	exp := now.Add(time.Duration(entity.AccessTokenTTLMinutes) * time.Minute)

	claims := &Claims{
		Email:         email,
		Roles:         roles,
		ActiveRole:    activeRole,
		KYCStatus:     kycStatus,
		EmailVerified: emailVerified,
		DeviceID:      deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{i.audience},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, claims, nil
}

// IssueRefreshToken creates a signed refresh token.
func (i *Issuer) IssueRefreshToken(userID, deviceID, familyID string) (string, string, error) {
	now := time.Now()
	jti := uuid.New().String()
	exp := now.Add(time.Duration(entity.RefreshTokenTTLDays) * 24 * time.Hour)

	claims := &jwt.RegisteredClaims{
		Issuer:    i.issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{i.audience},
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(i.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, jti, nil
}

// Verify validates a JWT access token and returns the claims.
func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// VerifyRefreshToken validates a refresh token and returns the registered claims.
func (v *Verifier) VerifyRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	return claims, nil
}

// GetPublicKey returns the public key from the issuer (for services that need both issue and verify).
func (i *Issuer) GetPublicKey() *rsa.PublicKey {
	return &i.privateKey.PublicKey
}
