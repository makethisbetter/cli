package project

import (
	"testing"
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
