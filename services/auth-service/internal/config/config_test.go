package config

import "testing"

func TestGetJWTKeyReturnsSecretBytes(t *testing.T) {
	cfg := &Config{
		JWTSecret: "super-secret",
	}

	got := string(cfg.GetJWTKey())
	if got != "super-secret" {
		t.Fatalf("GetJWTKey() = %q, want %q", got, "super-secret")
	}
}

