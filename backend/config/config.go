package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App              AppConfig              `mapstructure:"app"`
	Database         DatabaseConfig         `mapstructure:"database"`
	JWT              JWTConfig              `mapstructure:"jwt"`
	AA               AAConfig               `mapstructure:"aa"`
	Webhook          WebhookConfig          `mapstructure:"webhook"`
	SMTP             SMTPConfig             `mapstructure:"smtp"`
	BankVerification BankVerificationConfig `mapstructure:"bank_verification"`
}

type AppConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type AAConfig struct {
	BaseURL       string `mapstructure:"base_url"`
	APIKey        string `mapstructure:"api_key"`
	ClientID      string `mapstructure:"client_id"`
	ClientSecret  string `mapstructure:"client_secret"`
	EncPublicKey  string `mapstructure:"enc_public_key"`
	EncPrivateKey string `mapstructure:"enc_private_key"`
	Provider      string `mapstructure:"provider"`
}

type WebhookConfig struct {
	Secret string `mapstructure:"secret"`
}

type SMTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

type BankVerificationConfig struct {
	APIKey  string `mapstructure:"api_key"`
	APIURL  string `mapstructure:"api_url"`
	Enabled bool   `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.env", "development")

	// Ensure PORT environment variable is used if available
	if port := os.Getenv("PORT"); port != "" {
		viper.SetDefault("app.port", port)
	}

	// Database defaults
	viper.SetDefault("database.dsn", "postgres://postgres:postgres@localhost:5432/expensedb?sslmode=disable")

	// JWT defaults
	viper.SetDefault("jwt.secret", "replace-me-in-production")

	// AA defaults
	viper.SetDefault("aa.base_url", "https://sandbox.example-aa.com")
	viper.SetDefault("aa.provider", "mock")

	// Webhook defaults
	viper.SetDefault("webhook.secret", "replace-me-in-production")

	// SMTP defaults
	viper.SetDefault("smtp.host", "localhost")
	viper.SetDefault("smtp.port", 587)
	viper.SetDefault("smtp.user", "")
	viper.SetDefault("smtp.pass", "")

	// Bank Verification defaults
	viper.SetDefault("bank_verification.api_key", "")
	viper.SetDefault("bank_verification.api_url", "https://api.bankverification.com/v1/verify")
	viper.SetDefault("bank_verification.enabled", true)
}
