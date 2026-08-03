package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestLoginCompletesAcrossTwoCLIInvocations(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		requests []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.Path)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/agent_registration":
			json.NewEncoder(w).Encode(map[string]any{
				"registration_token": "registration-token-from-first-invocation",
				"expires_in":         300,
			})
		case "/api/v1/agent_registration/verify":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode verify request: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if got := body["registration_token"]; got != "registration-token-from-first-invocation" {
				t.Errorf("registration_token = %q, want token from first invocation", got)
			}
			if got := body["otp"]; got != "123456" {
				t.Errorf("otp = %q, want 123456", got)
			}
			json.NewEncoder(w).Encode(map[string]any{
				"user":      map[string]string{"id": "user_1", "email": "dev@example.com"},
				"account":   map[string]string{"id": "account_1", "name": "Development"},
				"api_token": map[string]string{"token": "api-token", "name": "Agent CLI"},
			})
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "config.json")
	first := runLoginCLI(t, configPath, "login", "--email", "dev@example.com", "--send-only", "--api-url", server.URL+"/api/v1")
	if !strings.Contains(first, "Verification code sent to dev@example.com") {
		t.Fatalf("first invocation output = %q", first)
	}

	second := runLoginCLI(t, configPath, "login", "--otp", "123456")
	if !strings.Contains(second, "Logged in as dev@example.com") {
		t.Fatalf("second invocation output = %q", second)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(configPath), "pending-login.json")); !os.IsNotExist(err) {
		t.Fatalf("pending login file still exists after successful login: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"/api/v1/agent_registration", "/api/v1/agent_registration/verify"}
	if !slices.Equal(requests, want) {
		t.Fatalf("request paths = %v, want %v", requests, want)
	}
}

func TestLoginOTPRequiresPendingLogin(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	output, err := runLoginCLIResult(t, configPath, "login", "--otp", "123456")
	if err == nil {
		t.Fatal("login succeeded without a pending login")
	}
	if !strings.Contains(output, "no pending login") {
		t.Fatalf("output = %q, want missing pending login error", output)
	}
}

func TestLoginRemovesExpiredPendingLoginWithoutVerifying(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		requests []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.Path)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"registration_token": "already-expired-token",
			"expires_in":         0,
		})
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "config.json")
	runLoginCLI(t, configPath, "login", "--email", "dev@example.com", "--send-only", "--api-url", server.URL+"/api/v1")

	output, err := runLoginCLIResult(t, configPath, "login", "--otp", "123456")
	if err == nil {
		t.Fatal("login succeeded with expired pending state")
	}
	if !strings.Contains(output, "pending login has expired") {
		t.Fatalf("output = %q, want expired pending login error", output)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(configPath), "pending-login.json")); !os.IsNotExist(err) {
		t.Fatalf("pending login file still exists after expiration: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	want := []string{"/api/v1/agent_registration"}
	if !slices.Equal(requests, want) {
		t.Fatalf("request paths = %v, want %v", requests, want)
	}
}

func runLoginCLI(t *testing.T, configPath string, args ...string) string {
	t.Helper()

	output, err := runLoginCLIResult(t, configPath, args...)
	if err != nil {
		t.Fatalf("makethisbetter %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return output
}

func runLoginCLIResult(t *testing.T, configPath string, args ...string) (string, error) {
	t.Helper()

	commandArgs := append([]string{"-test.run=TestLoginHelperProcess", "--"}, args...)
	command := exec.Command(os.Args[0], commandArgs...)
	env := slices.DeleteFunc(os.Environ(), func(value string) bool {
		return strings.HasPrefix(value, "GO_WANT_LOGIN_HELPER=") || strings.HasPrefix(value, "MAKETHISBETTER_CONFIG=")
	})
	command.Env = append(env, "GO_WANT_LOGIN_HELPER=1", "MAKETHISBETTER_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestLoginHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_LOGIN_HELPER") != "1" {
		return
	}

	separator := slices.Index(os.Args, "--")
	if separator == -1 {
		fmt.Fprintln(os.Stderr, "missing helper argument separator")
		os.Exit(2)
	}
	rootCmd.SetArgs(os.Args[separator+1:])
	if err := Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
