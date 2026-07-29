package project

import (
	"strings"
	"testing"
)

func TestUpdateRequiresArgument(t *testing.T) {
	cmd := updateCmd
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("update command should require an argument, but accepted none")
	}
}

func TestUpdateAcceptsSingleArgument(t *testing.T) {
	cmd := updateCmd
	err := cmd.Args(cmd, []string{"acme"})
	if err != nil {
		t.Errorf("update command should accept one argument, got error: %v", err)
	}
}

func TestUpdateRejectsTooManyArguments(t *testing.T) {
	cmd := updateCmd
	err := cmd.Args(cmd, []string{"acme", "extra"})
	if err == nil {
		t.Error("update command should reject more than one argument")
	}
}

func TestUpdateHasAIContextFlag(t *testing.T) {
	if updateCmd.Flags().Lookup("ai-context") == nil {
		t.Error("update command should expose an --ai-context flag")
	}
}

func TestUpdateWithoutFlagsFails(t *testing.T) {
	cmd := updateCmd
	err := runUpdate(cmd, []string{"acme"})
	if err == nil {
		t.Fatal("update without any flags should fail before hitting the network")
	}
	if got := err.Error(); !strings.Contains(got, "nothing to update") || !strings.Contains(got, "--ai-context") {
		t.Errorf("error should name the available flags, got %q", got)
	}
}

func TestCreateHasAIContextFlag(t *testing.T) {
	if createCmd.Flags().Lookup("ai-context") == nil {
		t.Error("create command should expose an --ai-context flag")
	}
}
