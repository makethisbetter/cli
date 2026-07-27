package project

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateRequiresArgument(t *testing.T) {
	cmd := createCmd
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("create command should require an argument, but accepted none")
	}
}

func TestCreateAcceptsSingleArgument(t *testing.T) {
	cmd := createCmd
	err := cmd.Args(cmd, []string{"Acme"})
	if err != nil {
		t.Errorf("create command should accept one argument, got error: %v", err)
	}
}

func TestCreateRejectsTooManyArguments(t *testing.T) {
	cmd := createCmd
	err := cmd.Args(cmd, []string{"Acme", "extra"})
	if err == nil {
		t.Error("create command should reject more than one argument")
	}
}

func TestCreateHasDomainFlag(t *testing.T) {
	if createCmd.Flags().Lookup("domain") == nil {
		t.Error("create command should expose a --domain flag")
	}
}

func TestCreateDomainFlagIsRequired(t *testing.T) {
	flag := createCmd.Flags().Lookup("domain")
	if flag == nil {
		t.Fatal("create command should expose a --domain flag")
	}
	if flag.Annotations[cobra.BashCompOneRequiredFlag] == nil {
		t.Error("--domain should be marked required so cobra refuses the command without it")
	}
	if !strings.Contains(flag.Usage, "required") {
		t.Errorf("--domain help text should say it is required, got %q", flag.Usage)
	}
}

func TestNormalizeDomainRejectsBadValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHint string
	}{
		{"empty", "", "--domain is required"},
		{"whitespace only", "   ", "--domain is required"},
		{"full url", "https://example.com/app", "--domain example.com"},
		{"host with path", "example.com/app", "no path or spaces"},
		{"no tld", "localhost", "including the TLD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDomain(tt.input)
			if err == nil {
				t.Fatalf("normalizeDomain(%q) should fail, got %q", tt.input, got)
			}
			if !strings.Contains(err.Error(), tt.wantHint) {
				t.Errorf("error should tell the user what to type: got %q, want it to contain %q", err.Error(), tt.wantHint)
			}
		})
	}
}

func TestNormalizeDomainAcceptsHostname(t *testing.T) {
	got, err := normalizeDomain("  app.example.com ")
	if err != nil {
		t.Fatalf("normalizeDomain should accept a hostname, got error: %v", err)
	}
	if got != "app.example.com" {
		t.Errorf("normalizeDomain should trim surrounding space: got %q, want %q", got, "app.example.com")
	}
}
