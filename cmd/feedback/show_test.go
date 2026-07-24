package feedback

import (
	"testing"
)

func TestShowRequiresArgument(t *testing.T) {
	cmd := showCmd
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("show command should require an argument, but accepted none")
	}
}

func TestShowAcceptsSingleArgument(t *testing.T) {
	cmd := showCmd
	err := cmd.Args(cmd, []string{"acme/FB-123"})
	if err != nil {
		t.Errorf("show command should accept one argument, got error: %v", err)
	}
}

func TestShowRejectsTooManyArguments(t *testing.T) {
	cmd := showCmd
	err := cmd.Args(cmd, []string{"acme/FB-1", "acme/FB-2"})
	if err == nil {
		t.Error("show command should reject more than one argument")
	}
}

func TestShowHasMarkdownFlag(t *testing.T) {
	if showCmd.Flags().Lookup("md") == nil {
		t.Error("show command should expose an --md flag for server-rendered markdown")
	}
}
