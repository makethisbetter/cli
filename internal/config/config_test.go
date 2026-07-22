package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFrom_FileNotFound(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.APIURL != DefaultAPIURL {
		t.Errorf("expected default API URL %q, got %q", DefaultAPIURL, cfg.APIURL)
	}
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %q", cfg.Token)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".makethisbetter", "config.json")

	original := &Config{
		Token:     "test-token-123",
		APIURL:    "https://example.com/api/v1/",
		AccountID: "acc_test",
		UserEmail: "test@example.com",
	}

	if err := SaveTo(original, path); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Token != "test-token-123" {
		t.Errorf("token: got %q, want %q", loaded.Token, "test-token-123")
	}
	if loaded.APIURL != "https://example.com/api/v1" {
		t.Errorf("api_url: got %q, want trailing slash stripped", loaded.APIURL)
	}
	if loaded.AccountID != "acc_test" {
		t.Errorf("account_id: got %q, want %q", loaded.AccountID, "acc_test")
	}
	if loaded.UserEmail != "test@example.com" {
		t.Errorf("user_email: got %q, want %q", loaded.UserEmail, "test@example.com")
	}
}

// The serialized token field name is a cross-tool contract: the MCP server and
// Skills read `api_token` from the same config file.
func TestSaveTo_SerializesAPITokenField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := SaveTo(&Config{Token: "test-token-123"}, path); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"api_token": "test-token-123"`) {
		t.Errorf("config JSON must use the api_token key, got: %s", data)
	}
	if strings.Contains(string(data), `"token"`) {
		t.Errorf("config JSON must not use the legacy token key, got: %s", data)
	}
}

func TestLoadFrom_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", DefaultAPIURL},
		{"https://example.com/api/v1", "https://example.com/api/v1"},
		{"https://example.com/api/v1/", "https://example.com/api/v1"},
		{"https://example.com/api/v1///", "https://example.com/api/v1"},
	}
	for _, tt := range tests {
		got := NormalizeURL(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRequireToken(t *testing.T) {
	cfg := &Config{Token: ""}
	_, err := RequireToken(cfg)
	if err == nil {
		t.Fatal("expected error for empty token")
	}

	cfg.Token = "abc"
	tok, err := RequireToken(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "abc" {
		t.Errorf("got %q, want %q", tok, "abc")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".makethisbetter", "config.json")

	cfg := &Config{Token: "secret", APIURL: DefaultAPIURL}
	if err := SaveTo(cfg, path); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions: got %o, want 600", perm)
	}
}

func TestEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := SaveTo(&Config{
		Token:  "file-token",
		APIURL: DefaultAPIURL,
	}, path); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MTB_API_TOKEN", "env-token")
	t.Setenv("MTB_API_URL", "https://custom.api/v1")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if cfg.Token != "env-token" {
		t.Errorf("token: got %q, want env override %q", cfg.Token, "env-token")
	}
	if cfg.APIURL != "https://custom.api/v1" {
		t.Errorf("api_url: got %q, want env override", cfg.APIURL)
	}
}
