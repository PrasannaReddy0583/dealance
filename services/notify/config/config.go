package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct{ App AppConfig; DB DBConfig; Redis RedisConfig; JWT JWTConfig }
type AppConfig struct{ Env, Port string }
type DBConfig struct{ Host, Port, User, Password, Name, SSLMode string }
func (c DBConfig) DSN() string { return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode) }
type RedisConfig struct{ Host, Port, Password string; DB int }
func (c RedisConfig) Addr() string { return c.Host + ":" + c.Port }
type JWTConfig struct{ PublicKeyPath, Issuer, Audience string }

func Load() (*Config, error) {
	viper.SetConfigFile(".env"); viper.AutomaticEnv(); _ = viper.ReadInConfig()
	viper.SetDefault("APP_ENV", "development"); viper.SetDefault("APP_PORT", "8088")
	viper.SetDefault("DB_HOST", "localhost"); viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "dealance"); viper.SetDefault("DB_PASSWORD", "dealance")
	viper.SetDefault("DB_NAME", "dealance_notify"); viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost"); viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", ""); viper.SetDefault("REDIS_DB", 8)
	viper.SetDefault("JWT_PUBLIC_KEY_PATH", "../../infrastructure/keys/public.pem")
	viper.SetDefault("JWT_ISSUER", "https://auth.dealance.com"); viper.SetDefault("JWT_AUDIENCE", "https://api.dealance.com")
	return &Config{
		App: AppConfig{Env: viper.GetString("APP_ENV"), Port: viper.GetString("APP_PORT")},
		DB: DBConfig{Host: viper.GetString("DB_HOST"), Port: viper.GetString("DB_PORT"), User: viper.GetString("DB_USER"), Password: viper.GetString("DB_PASSWORD"), Name: viper.GetString("DB_NAME"), SSLMode: viper.GetString("DB_SSL_MODE")},
		Redis: RedisConfig{Host: viper.GetString("REDIS_HOST"), Port: viper.GetString("REDIS_PORT"), Password: viper.GetString("REDIS_PASSWORD"), DB: viper.GetInt("REDIS_DB")},
		JWT: JWTConfig{PublicKeyPath: viper.GetString("JWT_PUBLIC_KEY_PATH"), Issuer: viper.GetString("JWT_ISSUER"), Audience: viper.GetString("JWT_AUDIENCE")},
	}, nil
}
