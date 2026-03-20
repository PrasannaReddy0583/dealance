package entity

// --- Signup DTOs ---

type InitiateSignupRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type InitiateSignupResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type VerifyEmailRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	OTP       string `json:"otp" validate:"required,len=6"`
}

type VerifyEmailResponse struct {
	UserID string `json:"user_id"`
	Stage  string `json:"stage"`
}

type ResendOTPRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

type ConfirmAuthRequest struct {
	UserID       string `json:"user_id" validate:"required,uuid"`
	ProviderType string `json:"provider_type" validate:"required,oneof=PASSKEY GOOGLE APPLE"`
	ExternalID   string `json:"external_id" validate:"required"`
	PublicKey    []byte `json:"public_key,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // For OAuth
}

type SetCountryRequest struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	CountryCode string `json:"country_code" validate:"required,len=2"`
}

type SetRoleRequest struct {
	UserID string   `json:"user_id" validate:"required,uuid"`
	Roles  []string `json:"roles" validate:"required,min=1,dive,oneof=ENTREPRENEUR INVESTOR"`
}

type InitiateKYCRequest struct {
	UserID   string `json:"user_id" validate:"required,uuid"`
	KYCType  string `json:"kyc_type" validate:"required,oneof=IDENTITY INVESTOR_ACCREDITATION"`
	Vendor   string `json:"vendor" validate:"required,oneof=HYPERVERGE ONFIDO"`
}

type InitiateKYCResponse struct {
	SessionID string `json:"session_id"`
	SDKToken  string `json:"sdk_token,omitempty"`
	Status    string `json:"status"`
}

type SignupStatusResponse struct {
	UserID        string   `json:"user_id"`
	Email         string   `json:"email"`
	CurrentStage  string   `json:"current_stage"`
	NextAction    string   `json:"next_action"`
	Roles         []string `json:"roles,omitempty"`
	KYCStatus     string   `json:"kyc_status,omitempty"`
	AccountStatus string   `json:"account_status"`
}

// --- Login DTOs ---

type BeginPasskeyLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type BeginPasskeyLoginResponse struct {
	ChallengeID string `json:"challenge_id"`
	Challenge   string `json:"challenge"` // base64 encoded 32-byte challenge
}

type FinishPasskeyLoginRequest struct {
	ChallengeID      string `json:"challenge_id" validate:"required"`
	CredentialID     string `json:"credential_id" validate:"required"`
	AuthenticatorData string `json:"authenticator_data" validate:"required"`
	ClientDataJSON   string `json:"client_data_json" validate:"required"`
	Signature        string `json:"signature" validate:"required"`
	DeviceID         string `json:"device_id,omitempty"`
}

type OAuthLoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=GOOGLE APPLE"`
	IDToken  string `json:"id_token" validate:"required"`
	DeviceID string `json:"device_id,omitempty"`
}

type BeginEmailLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type FinishEmailLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	OTP      string `json:"otp" validate:"required,len=6"`
	DeviceID string `json:"device_id,omitempty"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Roles         []string `json:"roles"`
	ActiveRole    string   `json:"active_role"`
	KYCStatus     string   `json:"kyc_status"`
	AccountStatus string   `json:"account_status"`
	SignupStage   string   `json:"signup_stage"`
}

// --- Token DTOs ---

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// --- Passkey Registration DTOs ---

type BeginPasskeyRegistrationRequest struct {
	DeviceName string `json:"device_name,omitempty"`
}

type BeginPasskeyRegistrationResponse struct {
	ChallengeID string `json:"challenge_id"`
	Challenge   string `json:"challenge"`
	UserID      string `json:"user_id"`
}

type FinishPasskeyRegistrationRequest struct {
	ChallengeID      string `json:"challenge_id" validate:"required"`
	CredentialID     string `json:"credential_id" validate:"required"`
	PublicKey        string `json:"public_key" validate:"required"` // base64
	AuthenticatorData string `json:"authenticator_data" validate:"required"`
	ClientDataJSON   string `json:"client_data_json" validate:"required"`
	Attestation      string `json:"attestation,omitempty"`
	DeviceName       string `json:"device_name,omitempty"`
}

// --- KYC Webhook DTOs ---

type HypervergeWebhookPayload struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	Result        struct {
		FaceMatch float64 `json:"faceMatch"`
		Liveness  float64 `json:"liveness"`
		Deepfake  float64 `json:"deepfake"`
	} `json:"result"`
	RejectionReason string `json:"rejectionReason,omitempty"`
}

type OnfidoWebhookPayload struct {
	Payload struct {
		ResourceType string `json:"resource_type"`
		Action       string `json:"action"`
		Object       struct {
			ID               string `json:"id"`
			Status           string `json:"status"`
			Result           string `json:"result"`
			SubResult        string `json:"sub_result,omitempty"`
		} `json:"object"`
	} `json:"payload"`
}
