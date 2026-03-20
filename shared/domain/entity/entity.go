package entity

// AccountStatus represents the lifecycle state of a user account.
type AccountStatus string

const (
	AccountStatusPending   AccountStatus = "PENDING"
	AccountStatusActive    AccountStatus = "ACTIVE"
	AccountStatusSuspended AccountStatus = "SUSPENDED"
	AccountStatusBanned    AccountStatus = "BANNED"
)

// SignupStage represents the current step in the multi-stage signup flow.
type SignupStage string

const (
	SignupStageEmailVerify SignupStage = "EMAIL_VERIFY"
	SignupStageAuthSetup   SignupStage = "AUTH_SETUP"
	SignupStageCountry     SignupStage = "COUNTRY"
	SignupStageRole        SignupStage = "ROLE"
	SignupStageKYC         SignupStage = "KYC"
	SignupStageComplete    SignupStage = "COMPLETE"
)

// SignupStageOrder defines the valid progression order for signup stages.
var SignupStageOrder = map[SignupStage]int{
	SignupStageEmailVerify: 0,
	SignupStageAuthSetup:   1,
	SignupStageCountry:     2,
	SignupStageRole:        3,
	SignupStageKYC:         4,
	SignupStageComplete:    5,
}

// CanAdvanceTo checks if advancing from the current stage to the target is valid.
func (s SignupStage) CanAdvanceTo(target SignupStage) bool {
	currentOrder, ok1 := SignupStageOrder[s]
	targetOrder, ok2 := SignupStageOrder[target]
	if !ok1 || !ok2 {
		return false
	}
	return targetOrder == currentOrder+1
}

// UserRole represents the possible roles a user can have.
type UserRole string

const (
	UserRoleEntrepreneur UserRole = "ENTREPRENEUR"
	UserRoleInvestor     UserRole = "INVESTOR"
	UserRoleBoth         UserRole = "BOTH"
	UserRoleAdmin        UserRole = "ADMIN"
)

// KYCStatus represents the verification status of a KYC record.
type KYCStatus string

const (
	KYCStatusPending       KYCStatus = "PENDING"
	KYCStatusApproved      KYCStatus = "APPROVED"
	KYCStatusRejected      KYCStatus = "REJECTED"
	KYCStatusRejectedFinal KYCStatus = "REJECTED_FINAL"
)

// KYCType represents the type of KYC verification.
type KYCType string

const (
	KYCTypeIdentity               KYCType = "IDENTITY"
	KYCTypeInvestorAccreditation   KYCType = "INVESTOR_ACCREDITATION"
)

// KYCVendor represents the KYC verification provider.
type KYCVendor string

const (
	KYCVendorHyperverge KYCVendor = "HYPERVERGE"
	KYCVendorOnfido     KYCVendor = "ONFIDO"
	KYCVendorManual     KYCVendor = "MANUAL"
)

// DocumentType for KYC document verification.
type DocumentType string

const (
	DocumentTypeAadhaar  DocumentType = "AADHAAR"
	DocumentTypePassport DocumentType = "PASSPORT"
	DocumentTypePAN      DocumentType = "PAN"
)

// ProviderType for identity providers.
type ProviderType string

const (
	ProviderTypePasskey ProviderType = "PASSKEY"
	ProviderTypeGoogle  ProviderType = "GOOGLE"
	ProviderTypeApple   ProviderType = "APPLE"
)

// Platform represents the device platform.
type Platform string

const (
	PlatformIOS     Platform = "IOS"
	PlatformAndroid Platform = "ANDROID"
)

// RiskLevel for device attestation.
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "LOW"
	RiskLevelMedium RiskLevel = "MEDIUM"
	RiskLevelHigh   RiskLevel = "HIGH"
)

// ContentType represents the type of content.
type ContentType string

const (
	ContentTypePost  ContentType = "POST"
	ContentTypeShort ContentType = "SHORT"
	ContentTypeVideo ContentType = "VIDEO"
)

// InvestmentType represents the type of investment.
type InvestmentType string

const (
	InvestmentTypeEquity          InvestmentType = "EQUITY"
	InvestmentTypeLoan            InvestmentType = "LOAN"
	InvestmentTypeRoyalty         InvestmentType = "ROYALTY"
	InvestmentTypeConvertibleNote InvestmentType = "CONVERTIBLE_NOTE"
	InvestmentTypeSAFE            InvestmentType = "SAFE"
)

// InvestmentStatus represents the lifecycle of an investment.
type InvestmentStatus string

const (
	InvestmentStatusCommitted   InvestmentStatus = "COMMITTED"
	InvestmentStatusDeposited   InvestmentStatus = "DEPOSITED"
	InvestmentStatusInEscrow    InvestmentStatus = "IN_ESCROW"
	InvestmentStatusTransferred InvestmentStatus = "TRANSFERRED"
	InvestmentStatusCompleted   InvestmentStatus = "COMPLETED"
	InvestmentStatusRefunded    InvestmentStatus = "REFUNDED"
)

// ConversationType for chat service.
type ConversationType string

const (
	ConversationTypeDirect      ConversationType = "DIRECT"
	ConversationTypeGroup       ConversationType = "GROUP"
	ConversationTypeDealRoom    ConversationType = "DEAL_ROOM"
	ConversationTypeStartupTeam ConversationType = "STARTUP_TEAM"
)

// VerificationStatus for startups.
type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "UNVERIFIED"
	VerificationStatusPending    VerificationStatus = "PENDING"
	VerificationStatusVerified   VerificationStatus = "VERIFIED"
	VerificationStatusRejected   VerificationStatus = "REJECTED"
)

// InvestorType for investor profiles.
type InvestorType string

const (
	InvestorTypeAngel        InvestorType = "ANGEL"
	InvestorTypeVC           InvestorType = "VC"
	InvestorTypeInstitutional InvestorType = "INSTITUTIONAL"
	InvestorTypeHNI          InvestorType = "HNI"
	InvestorTypeFamilyOffice InvestorType = "FAMILY_OFFICE"
)

// StartupStage represents the stage of a startup.
type StartupStage string

const (
	StartupStageIdea     StartupStage = "IDEA"
	StartupStagePreSeed  StartupStage = "PRE_SEED"
	StartupStageSeed     StartupStage = "SEED"
	StartupStageSeriesA  StartupStage = "SERIES_A"
	StartupStageSeriesB  StartupStage = "SERIES_B"
	StartupStageGrowth   StartupStage = "GROWTH"
)

// Depository for investor verifications.
type Depository string

const (
	DepositoryNSDL Depository = "NSDL"
	DepositoryCDSL Depository = "CDSL"
)

// KYC Thresholds
const (
	KYCFaceMatchThreshold  = 0.90
	KYCLivenessThreshold   = 0.85
	KYCDeepfakeMaxThreshold = 0.15
	KYCMaxAttempts         = 3
)

// SEBI Accreditation Thresholds (in paise)
const (
	SEBINetWorthThresholdPaise    int64 = 2_00_00_000_00 // 2 Crore INR = 20,000,000 paise
	SEBIAnnualIncomeThresholdPaise int64 = 25_00_000_00   // 25 Lakh INR = 2,500,000 paise
)

// Token TTLs
const (
	AccessTokenTTLMinutes  = 15
	RefreshTokenTTLDays    = 30
)

// AuditEventType captures the type of security audit event.
type AuditEventType string

const (
	AuditEventSignupInitiated   AuditEventType = "SIGNUP_INITIATED"
	AuditEventEmailVerified     AuditEventType = "EMAIL_VERIFIED"
	AuditEventLoginSuccess      AuditEventType = "LOGIN_SUCCESS"
	AuditEventLoginFailed       AuditEventType = "LOGIN_FAILED"
	AuditEventTokenRefreshed    AuditEventType = "TOKEN_REFRESHED"
	AuditEventLogout            AuditEventType = "LOGOUT"
	AuditEventPasskeyRegistered AuditEventType = "PASSKEY_REGISTERED"
	AuditEventKYCInitiated      AuditEventType = "KYC_INITIATED"
	AuditEventKYCApproved       AuditEventType = "KYC_APPROVED"
	AuditEventKYCRejected       AuditEventType = "KYC_REJECTED"
	AuditEventDeviceAttested    AuditEventType = "DEVICE_ATTESTED"
	AuditEventReplayDetected    AuditEventType = "REPLAY_DETECTED"
	AuditEventRateLimited       AuditEventType = "RATE_LIMITED"
)
