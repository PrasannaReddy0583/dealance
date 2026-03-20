package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the auth service.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	Scylla   ScyllaConfig
	JWT      JWTConfig
	KYC      KYCConfig
	OAuth    OAuthConfig
	Email    EmailConfig
}

type AppConfig struct {
	Env          string `mapstructure:"APP_ENV"`
	Port         string `mapstructure:"APP_PORT"`
	SkipAttest   bool   `mapstructure:"SKIP_ATTEST"`
	SkipSigning  bool   `mapstructure:"SKIP_SIGNING"`
	KYCMock      bool   `mapstructure:"KYC_MOCK"`
}

type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type ScyllaConfig struct {
	Hosts    string `mapstructure:"SCYLLA_HOSTS"`
	Keyspace string `mapstructure:"SCYLLA_KEYSPACE"`
}

func (c ScyllaConfig) HostList() []string {
	return strings.Split(c.Hosts, ",")
}

type JWTConfig struct {
	PrivateKeyPath string `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	PublicKeyPath  string `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	Issuer         string `mapstructure:"JWT_ISSUER"`
	Audience       string `mapstructure:"JWT_AUDIENCE"`
}

type KYCConfig struct {
	HypervergeAPIKey    string `mapstructure:"HYPERVERGE_API_KEY"`
	HypervergeSecret    string `mapstructure:"HYPERVERGE_SECRET"`
	HypervergeBaseURL   string `mapstructure:"HYPERVERGE_BASE_URL"`
	OnfidoAPIToken      string `mapstructure:"ONFIDO_API_TOKEN"`
	OnfidoWebhookSecret string `mapstructure:"ONFIDO_WEBHOOK_SECRET"`
}

type OAuthConfig struct {
	GoogleClientID  string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleJWKSURL   string `mapstructure:"GOOGLE_JWKS_URL"`
	AppleClientID   string `mapstructure:"APPLE_CLIENT_ID"`
	AppleTeamID     string `mapstructure:"APPLE_TEAM_ID"`
	AppleJWKSURL    string `mapstructure:"APPLE_JWKS_URL"`
}

type EmailConfig struct {
	SMTPHost string `mapstructure:"SMTP_HOST"`
	SMTPPort string `mapstructure:"SMTP_PORT"`
	FromAddr string `mapstructure:"EMAIL_FROM"`
}

// Load reads configuration from env vars and .env file.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Try to read .env file (not required)
	_ = viper.ReadInConfig()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("SKIP_ATTEST", true)
	viper.SetDefault("SKIP_SIGNING", true)
	viper.SetDefault("KYC_MOCK", true)

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "dealance")
	viper.SetDefault("DB_PASSWORD", "dealance")
	viper.SetDefault("DB_NAME", "dealance_auth")
	viper.SetDefault("DB_SSL_MODE", "disable")

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("SCYLLA_HOSTS", "localhost:9042")
	viper.SetDefault("SCYLLA_KEYSPACE", "dealance_auth")

	viper.SetDefault("JWT_PRIVATE_KEY_PATH", "./keys/private.pem")
	viper.SetDefault("JWT_PUBLIC_KEY_PATH", "./keys/public.pem")
	viper.SetDefault("JWT_ISSUER", "https://auth.dealance.com")
	viper.SetDefault("JWT_AUDIENCE", "https://api.dealance.com")

	viper.SetDefault("GOOGLE_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs")
	viper.SetDefault("APPLE_JWKS_URL", "https://appleid.apple.com/auth/keys")

	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_PORT", "1025")
	viper.SetDefault("EMAIL_FROM", "noreply@dealance.com")

	cfg := &Config{
		App: AppConfig{
			Env:         viper.GetString("APP_ENV"),
			Port:        viper.GetString("APP_PORT"),
			SkipAttest:  viper.GetBool("SKIP_ATTEST"),
			SkipSigning: viper.GetBool("SKIP_SIGNING"),
			KYCMock:     viper.GetBool("KYC_MOCK"),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Scylla: ScyllaConfig{
			Hosts:    viper.GetString("SCYLLA_HOSTS"),
			Keyspace: viper.GetString("SCYLLA_KEYSPACE"),
		},
		JWT: JWTConfig{
			PrivateKeyPath: viper.GetString("JWT_PRIVATE_KEY_PATH"),
			PublicKeyPath:  viper.GetString("JWT_PUBLIC_KEY_PATH"),
			Issuer:         viper.GetString("JWT_ISSUER"),
			Audience:       viper.GetString("JWT_AUDIENCE"),
		},
		KYC: KYCConfig{
			HypervergeAPIKey:    viper.GetString("HYPERVERGE_API_KEY"),
			HypervergeSecret:    viper.GetString("HYPERVERGE_SECRET"),
			HypervergeBaseURL:   viper.GetString("HYPERVERGE_BASE_URL"),
			OnfidoAPIToken:      viper.GetString("ONFIDO_API_TOKEN"),
			OnfidoWebhookSecret: viper.GetString("ONFIDO_WEBHOOK_SECRET"),
		},
		OAuth: OAuthConfig{
			GoogleClientID: viper.GetString("GOOGLE_CLIENT_ID"),
			GoogleJWKSURL:  viper.GetString("GOOGLE_JWKS_URL"),
			AppleClientID:  viper.GetString("APPLE_CLIENT_ID"),
			AppleTeamID:    viper.GetString("APPLE_TEAM_ID"),
			AppleJWKSURL:   viper.GetString("APPLE_JWKS_URL"),
		},
		Email: EmailConfig{
			SMTPHost: viper.GetString("SMTP_HOST"),
			SMTPPort: viper.GetString("SMTP_PORT"),
			FromAddr: viper.GetString("EMAIL_FROM"),
		},
	}

	return cfg, nil
}
