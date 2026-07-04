package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultAPIURL is the production API base URL used when none is configured.
const DefaultAPIURL = "https://makethisbetter.dev/api/v1"

// Config holds the persisted CLI credentials and API settings.
type Config struct {
	Token     string `json:"api_token,omitempty"`
	APIURL    string `json:"api_url"`
	AccountID string `json:"account_id,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

// DefaultPath returns the default config file location under the user's home.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".makethisbetter", "config.json"), nil
}

// Load reads the config from the default path, applying environment overrides.
func Load() (*Config, error) {
	return LoadFrom("")
}

// LoadFrom reads the config from path (or the default path when empty),
// returning a default config if the file does not exist.
func LoadFrom(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON at %s: %w", path, err)
	}

	cfg.APIURL = NormalizeURL(cfg.APIURL)
	result := applyEnvOverrides(cfg)
	return result, nil
}

// Save writes cfg to the default config path.
func Save(cfg *Config) error {
	return SaveTo(cfg, "")
}

// SaveTo writes cfg to path (or the default path when empty), creating the
// containing directory with owner-only permissions.
func SaveTo(cfg *Config, path string) error {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return err
		}
	}

	cfg.APIURL = NormalizeURL(cfg.APIURL)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ErrNotLoggedIn is returned when an API call requires authentication
// but no token is configured.
var ErrNotLoggedIn = errors.New("not logged in, run `makethisbetter login` first")

// RequireToken returns the configured token or ErrNotLoggedIn when it is unset.
func RequireToken(cfg *Config) (string, error) {
	if cfg.Token == "" {
		return "", ErrNotLoggedIn
	}
	return cfg.Token, nil
}

// NormalizeURL trims a trailing slash and falls back to DefaultAPIURL when empty.
func NormalizeURL(u string) string {
	if u == "" {
		return DefaultAPIURL
	}
	return strings.TrimRight(u, "/")
}

func defaultConfig() *Config {
	cfg := &Config{APIURL: DefaultAPIURL}
	return applyEnvOverrides(*cfg)
}

func applyEnvOverrides(cfg Config) *Config {
	if v := os.Getenv("MTB_API_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("MTB_API_URL"); v != "" {
		cfg.APIURL = NormalizeURL(v)
	}
	return &cfg
}
