package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/auth/internal/domain/entity"
	autherrors "github.com/dealance/services/auth/internal/domain/errors"
	"github.com/dealance/services/auth/internal/domain/repository"
	sharedentity "github.com/dealance/shared/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/crypto"
)

// SignupService handles multi-stage user registration.
type SignupService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.UserRoleRepository
	identRepo    repository.IdentityProviderRepository
	kycRepo      repository.KYCRepository
	sessionRepo  repository.SessionRepository
	auditRepo    repository.AuditLogRepository
	emailSvc     repository.EmailService
	kycVendorSvc repository.KYCVendorService
	log          zerolog.Logger
	kycMock      bool
}

// NewSignupService creates a new signup service.
func NewSignupService(
	userRepo repository.UserRepository,
	roleRepo repository.UserRoleRepository,
	identRepo repository.IdentityProviderRepository,
	kycRepo repository.KYCRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
	emailSvc repository.EmailService,
	kycVendorSvc repository.KYCVendorService,
	log zerolog.Logger,
	kycMock bool,
) *SignupService {
	return &SignupService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		identRepo:    identRepo,
		kycRepo:      kycRepo,
		sessionRepo:  sessionRepo,
		auditRepo:    auditRepo,
		emailSvc:     emailSvc,
		kycVendorSvc: kycVendorSvc,
		log:          log,
		kycMock:      kycMock,
	}
}

// Initiate starts the signup flow — validates email, sends OTP, creates Redis session.
func (s *SignupService) Initiate(ctx context.Context, req entity.InitiateSignupRequest) (*entity.InitiateSignupResponse, error) {
	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	if exists {
		return nil, autherrors.ErrEmailAlreadyExists
	}

	// Generate OTP
	otp, err := crypto.GenerateOTP()
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Hash OTP for storage
	otpHash := crypto.HashOTP(otp)

	// Create session in Redis (TTL 10 minutes)
	sessionID := uuid.New().String()
	err = s.sessionRepo.CreateSignupSession(ctx, sessionID, req.Email, otpHash, 10*time.Minute)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Send OTP via email
	err = s.emailSvc.SendOTP(ctx, req.Email, otp)
	if err != nil {
		s.log.Error().Err(err).Str("email", req.Email).Msg("failed to send OTP email")
		// Don't fail the request — the OTP is stored, user can resend
	}

	s.log.Info().Str("email", req.Email).Str("session_id", sessionID).Msg("signup initiated")

	return &entity.InitiateSignupResponse{
		SessionID: sessionID,
		Message:   "Verification code sent to your email",
	}, nil
}

// VerifyEmail verifies the OTP, creates the user row, and advances to AUTH_SETUP.
func (s *SignupService) VerifyEmail(ctx context.Context, req entity.VerifyEmailRequest) (*entity.VerifyEmailResponse, error) {
	// Get session from Redis
	email, otpHash, err := s.sessionRepo.GetSignupSession(ctx, req.SessionID)
	if err != nil {
		return nil, autherrors.ErrSessionNotFound
	}

	// Verify OTP
	if !crypto.VerifyOTP(req.OTP, otpHash) {
		return nil, apperrors.ErrOTPInvalid
	}

	// Check if email was registered between initiate and verify
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	if exists {
		return nil, autherrors.ErrEmailAlreadyExists
	}

	// Create user
	user, err := s.userRepo.Create(ctx, email)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Mark email as verified
	err = s.userRepo.UpdateEmailVerified(ctx, user.ID)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Advance stage to AUTH_SETUP
	err = s.userRepo.UpdateSignupStage(ctx, user.ID, string(sharedentity.SignupStageAuthSetup))
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Delete session
	_ = s.sessionRepo.DeleteSignupSession(ctx, req.SessionID)

	// Audit log
	s.logAudit(ctx, user.ID, string(sharedentity.AuditEventEmailVerified), "", "")

	return &entity.VerifyEmailResponse{
		UserID: user.ID.String(),
		Stage:  string(sharedentity.SignupStageAuthSetup),
	}, nil
}

// ResendOTP generates a new OTP for the existing signup session.
func (s *SignupService) ResendOTP(ctx context.Context, req entity.ResendOTPRequest) error {
	// Get session
	email, _, err := s.sessionRepo.GetSignupSession(ctx, req.SessionID)
	if err != nil {
		return autherrors.ErrSessionNotFound
	}

	// Generate new OTP
	otp, err := crypto.GenerateOTP()
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	otpHash := crypto.HashOTP(otp)

	// Update session with new OTP hash
	err = s.sessionRepo.UpdateSignupSessionOTP(ctx, req.SessionID, otpHash)
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Send new OTP
	err = s.emailSvc.SendOTP(ctx, email, otp)
	if err != nil {
		s.log.Error().Err(err).Str("email", email).Msg("failed to resend OTP")
	}

	return nil
}

// ConfirmAuth confirms passkey or OAuth provider link and advances to COUNTRY stage.
func (s *SignupService) ConfirmAuth(ctx context.Context, req entity.ConfirmAuthRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return apperrors.ErrValidation("Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return autherrors.ErrUserNotFound
	}

	// Validate stage
	if user.SignupStage != string(sharedentity.SignupStageAuthSetup) {
		return autherrors.ErrStageInvalid
	}

	// Create identity provider record
	ip := &entity.IdentityProvider{
		ID:           uuid.New(),
		UserID:       userID,
		ProviderType: req.ProviderType,
		ExternalID:   req.ExternalID,
		PublicKey:    req.PublicKey,
		DeviceName:   sql.NullString{String: req.DeviceName, Valid: req.DeviceName != ""},
	}

	err = s.identRepo.Create(ctx, ip)
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Advance to COUNTRY stage
	err = s.userRepo.UpdateSignupStage(ctx, userID, string(sharedentity.SignupStageCountry))
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	return nil
}

// SetCountry validates the country and advances to ROLE stage.
func (s *SignupService) SetCountry(ctx context.Context, req entity.SetCountryRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return apperrors.ErrValidation("Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return autherrors.ErrUserNotFound
	}

	if user.SignupStage != string(sharedentity.SignupStageCountry) {
		return autherrors.ErrStageInvalid
	}

	// OFAC sanctions check (simplified — in production, call an actual API)
	if isOFACSanctioned(req.CountryCode) {
		return autherrors.ErrSanctionsMatch
	}

	// Supported country check
	if !isSupportedCountry(req.CountryCode) {
		return autherrors.ErrCountryNotSupported
	}

	err = s.userRepo.UpdateCountryCode(ctx, userID, req.CountryCode)
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	err = s.userRepo.UpdateSignupStage(ctx, userID, string(sharedentity.SignupStageRole))
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	return nil
}

// SetRole sets user roles and advances to KYC stage.
// Auto-advances through AUTH_SETUP and COUNTRY stages if the frontend skips them.
func (s *SignupService) SetRole(ctx context.Context, req entity.SetRoleRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return apperrors.ErrValidation("Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return autherrors.ErrUserNotFound
	}

	// Auto-advance through intermediate stages that the frontend skips
	currentStage := user.SignupStage
	allowedStages := map[string]bool{
		string(sharedentity.SignupStageAuthSetup): true,
		string(sharedentity.SignupStageCountry):   true,
		string(sharedentity.SignupStageRole):      true,
	}

	if !allowedStages[currentStage] {
		return autherrors.ErrStageInvalid
	}

	// If at AUTH_SETUP, advance to COUNTRY first
	if currentStage == string(sharedentity.SignupStageAuthSetup) {
		_ = s.userRepo.UpdateSignupStage(ctx, userID, string(sharedentity.SignupStageCountry))
		currentStage = string(sharedentity.SignupStageCountry)
	}

	// If at COUNTRY, set a default country (IN) and advance to ROLE
	if currentStage == string(sharedentity.SignupStageCountry) {
		_ = s.userRepo.UpdateCountryCode(ctx, userID, "IN")
		_ = s.userRepo.UpdateSignupStage(ctx, userID, string(sharedentity.SignupStageRole))
	}

	for _, role := range req.Roles {
		err = s.roleRepo.Create(ctx, userID, role)
		if err != nil {
			return apperrors.ErrInternal().WithInternal(err)
		}
	}

	err = s.userRepo.UpdateSignupStage(ctx, userID, string(sharedentity.SignupStageKYC))
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	return nil
}

// InitiateKYC starts KYC verification and returns SDK token.
func (s *SignupService) InitiateKYC(ctx context.Context, req entity.InitiateKYCRequest) (*entity.InitiateKYCResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, apperrors.ErrValidation("Invalid user ID")
	}

	// Check attempt count
	attempts, err := s.kycRepo.CountAttempts(ctx, userID, req.KYCType)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	if attempts >= sharedentity.KYCMaxAttempts {
		return nil, autherrors.ErrKYCMaxAttempts
	}

	// Check if already approved
	latest, err := s.kycRepo.GetLatestByUserIDAndType(ctx, userID, req.KYCType)
	if err == nil && latest != nil && latest.Status == string(sharedentity.KYCStatusApproved) {
		return nil, autherrors.ErrKYCAlreadyApproved
	}

	// Create KYC record
	kycRecord := &entity.KYCRecord{
		ID:            uuid.New(),
		UserID:        userID,
		KYCType:       req.KYCType,
		Vendor:        sql.NullString{String: req.Vendor, Valid: true},
		Status:        string(sharedentity.KYCStatusPending),
		AttemptNumber: attempts + 1,
		CreatedAt:     time.Now(),
	}

	err = s.kycRepo.Create(ctx, kycRecord)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Call vendor API (or mock)
	var sessionID, sdkToken string
	if s.kycMock {
		sessionID = fmt.Sprintf("mock_%s", uuid.New().String()[:8])
		sdkToken = fmt.Sprintf("mock_sdk_token_%s", uuid.New().String()[:8])
	} else {
		sessionID, sdkToken, err = s.kycVendorSvc.InitiateSession(ctx, req.UserID, req.KYCType)
		if err != nil {
			return nil, apperrors.ErrInternal().WithInternal(err)
		}
	}

	// Update vendor session ID
	err = s.kycRepo.UpdateVendorSession(ctx, kycRecord.ID, sessionID)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Audit log
	s.logAudit(ctx, userID, string(sharedentity.AuditEventKYCInitiated), "", "")

	return &entity.InitiateKYCResponse{
		SessionID: sessionID,
		SDKToken:  sdkToken,
		Status:    string(sharedentity.KYCStatusPending),
	}, nil
}

// GetSignupStatus returns the current signup stage and what's needed next.
func (s *SignupService) GetSignupStatus(ctx context.Context, userID string) (*entity.SignupStatusResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("Invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, autherrors.ErrUserNotFound
	}

	roles, err := s.roleRepo.GetActiveRoles(ctx, uid)
	if err != nil {
		roles = []string{}
	}

	var kycStatus string
	latest, err := s.kycRepo.GetLatestByUserIDAndType(ctx, uid, string(sharedentity.KYCTypeIdentity))
	if err == nil && latest != nil {
		kycStatus = latest.Status
	}

	nextAction := getNextAction(user.SignupStage)

	return &entity.SignupStatusResponse{
		UserID:        userID,
		Email:         user.Email,
		CurrentStage:  user.SignupStage,
		NextAction:    nextAction,
		Roles:         roles,
		KYCStatus:     kycStatus,
		AccountStatus: user.AccountStatus,
	}, nil
}

func (s *SignupService) logAudit(ctx context.Context, userID uuid.UUID, eventType, deviceID, ipAddress string) {
	entry := &entity.AuditLogEntry{
		UserID:    userID,
		EventAt:   time.Now(),
		EventID:   uuid.New(),
		DeviceID:  deviceID,
		EventType: eventType,
		IPAddress: ipAddress,
		RiskScore: 0.0,
	}
	if err := s.auditRepo.Log(ctx, entry); err != nil {
		s.log.Error().Err(err).Str("event_type", eventType).Msg("failed to write audit log")
	}
}

// --- Helper functions ---

func getNextAction(stage string) string {
	switch stage {
	case string(sharedentity.SignupStageEmailVerify):
		return "Verify your email with the OTP sent"
	case string(sharedentity.SignupStageAuthSetup):
		return "Set up passkey or link OAuth provider"
	case string(sharedentity.SignupStageCountry):
		return "Select your country"
	case string(sharedentity.SignupStageRole):
		return "Choose your role (Entrepreneur/Investor)"
	case string(sharedentity.SignupStageKYC):
		return "Complete KYC verification"
	case string(sharedentity.SignupStageComplete):
		return "Signup complete — you can now log in"
	default:
		return "Unknown stage"
	}
}

// OFAC sanctioned countries (simplified list)
var sanctionedCountries = map[string]bool{
	"KP": true, // North Korea
	"IR": true, // Iran
	"SY": true, // Syria
	"CU": true, // Cuba
	"SD": true, // Sudan
}

func isOFACSanctioned(countryCode string) bool {
	return sanctionedCountries[countryCode]
}

// Supported countries for Dealance (India-focused)
var supportedCountries = map[string]bool{
	"IN": true,
	"US": true,
	"GB": true,
	"SG": true,
	"AE": true,
}

func isSupportedCountry(countryCode string) bool {
	return supportedCountries[countryCode]
}
