package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
}

type AppConfig struct {
	Env  string `mapstructure:"APP_ENV"`
	Port string `mapstructure:"APP_PORT"`
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

type JWTConfig struct {
	PublicKeyPath string `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	Issuer        string `mapstructure:"JWT_ISSUER"`
	Audience      string `mapstructure:"JWT_AUDIENCE"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8082")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "dealance")
	viper.SetDefault("DB_PASSWORD", "dealance")
	viper.SetDefault("DB_NAME", "dealance_content")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 2)
	viper.SetDefault("JWT_PUBLIC_KEY_PATH", "../../infrastructure/keys/public.pem")
	viper.SetDefault("JWT_ISSUER", "https://auth.dealance.com")
	viper.SetDefault("JWT_AUDIENCE", "https://api.dealance.com")

	cfg := &Config{
		App:   AppConfig{Env: viper.GetString("APP_ENV"), Port: viper.GetString("APP_PORT")},
		DB:    DBConfig{Host: viper.GetString("DB_HOST"), Port: viper.GetString("DB_PORT"), User: viper.GetString("DB_USER"), Password: viper.GetString("DB_PASSWORD"), Name: viper.GetString("DB_NAME"), SSLMode: viper.GetString("DB_SSL_MODE")},
		Redis: RedisConfig{Host: viper.GetString("REDIS_HOST"), Port: viper.GetString("REDIS_PORT"), Password: viper.GetString("REDIS_PASSWORD"), DB: viper.GetInt("REDIS_DB")},
		JWT:   JWTConfig{PublicKeyPath: viper.GetString("JWT_PUBLIC_KEY_PATH"), Issuer: viper.GetString("JWT_ISSUER"), Audience: viper.GetString("JWT_AUDIENCE")},
	}
	return cfg, nil
}
