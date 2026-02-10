package config

import "testing"

func TestGetEnvReturnsDefaultWhenUnset(t *testing.T) {
	const key = "K8S_CONFIG_TEST_UNSET"

	t.Setenv(key, "")

	got := getEnv(key, "default-value")
	if got != "default-value" {
		t.Fatalf("getEnv(%q) = %q, want %q", key, got, "default-value")
	}
}

func TestGetEnvReturnsValueWhenSet(t *testing.T) {
	const key = "K8S_CONFIG_TEST_SET"

	t.Setenv(key, "from-env")

	got := getEnv(key, "default-value")
	if got != "from-env" {
		t.Fatalf("getEnv(%q) = %q, want %q", key, got, "from-env")
	}
}

