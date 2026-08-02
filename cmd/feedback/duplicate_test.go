package feedback

import "testing"

func TestValidateCanonicalReference(t *testing.T) {
	if err := validateCanonicalReference("acme/FB-2", "acme/FB-1"); err != nil {
		t.Fatalf("same-project canonical should be accepted: %v", err)
	}
	if err := validateCanonicalReference("acme/FB-2", ""); err == nil {
		t.Fatal("duplicate should require canonical feedback")
	}
	if err := validateCanonicalReference("acme/FB-2", "FB-1"); err == nil {
		t.Fatal("bare canonical reference should be rejected")
	}
	if err := validateCanonicalReference("acme/FB-2", "other/FB-1"); err == nil {
		t.Fatal("cross-project canonical reference should be rejected")
	}
}

func TestDuplicateHasCanonicalFlag(t *testing.T) {
	if duplicateCmd.Flags().Lookup("canonical") == nil {
		t.Fatal("duplicate should expose a --canonical flag")
	}
}
