package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading config with default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default values
	if cfg.App.Port == "" {
		t.Error("Expected default port to be set")
	}

	if cfg.App.Env == "" {
		t.Error("Expected default env to be set")
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "test")
	os.Setenv("DATABASE_DSN", "postgres://test:test@localhost:5432/testdb")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	// Check that environment variables are respected
	if cfg.App.Port != "9090" {
		t.Errorf("Expected port to be 9090, got %s", cfg.App.Port)
	}

	if cfg.App.Env != "test" {
		t.Errorf("Expected env to be test, got %s", cfg.App.Env)
	}

	// Clean up
	os.Unsetenv("APP_PORT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DATABASE_DSN")
} 