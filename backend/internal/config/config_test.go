package config

import "testing"

func TestLoadRejectsPlaceholderJWTSecret(t *testing.T) {
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("JWT_SECRET", "replace_with_a_long_random_secret_at_least_32_characters")
	if _, err := Load(); err == nil {
		t.Fatal("placeholder secret must be rejected")
	}
	t.Setenv("JWT_SECRET", "k3J9vQ2mZpX8aLw0TnB5yRcD7eHfUg1S")
	if _, err := Load(); err != nil {
		t.Fatalf("real secret should load: %v", err)
	}
}