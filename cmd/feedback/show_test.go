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
	err := cmd.Args(cmd, []string{"fb_123"})
	if err != nil {
		t.Errorf("show command should accept one argument, got error: %v", err)
	}
}

func TestShowRejectsTooManyArguments(t *testing.T) {
	cmd := showCmd
	err := cmd.Args(cmd, []string{"fb_1", "fb_2"})
	if err == nil {
		t.Error("show command should reject more than one argument")
	}
}
