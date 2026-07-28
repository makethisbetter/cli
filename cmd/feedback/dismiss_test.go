package feedback

import (
	"strings"
	"testing"
)

func TestValidateDismissOptionsRequiresCanonicalForDuplicate(t *testing.T) {
	if err := validateDismissOptions("acme/FB-1", "duplicate", ""); err == nil {
		t.Fatal("duplicate should require canonical feedback")
	}
	if err := validateDismissOptions("acme/FB-1", "duplicate", "acme/FB-2"); err != nil {
		t.Fatalf("duplicate should accept canonical feedback: %v", err)
	}
}

func TestValidateDismissOptionsAllowsNotPlannedWithoutCanonical(t *testing.T) {
	if err := validateDismissOptions("acme/FB-1", "not_planned", ""); err != nil {
		t.Fatalf("not_planned should not require canonical feedback: %v", err)
	}
}

// A bare FB-n is what --json prints as `id`, so it is the shorthand users type.
// The server cannot resolve it, so the error must name the qualified form.
func TestValidateDismissOptionsRejectsBareCanonicalNumber(t *testing.T) {
	err := validateDismissOptions("acme/FB-2", "duplicate", "FB-1")
	if err == nil {
		t.Fatal("bare canonical number should be rejected locally")
	}
	if !strings.Contains(err.Error(), "acme/FB-1") {
		t.Errorf("error should name the qualified form, got %q", err)
	}
}

func TestValidateDismissOptionsRejectsCrossProjectCanonical(t *testing.T) {
	err := validateDismissOptions("acme/FB-2", "duplicate", "other/FB-1")
	if err == nil {
		t.Fatal("canonical from another project should be rejected locally")
	}
	if !strings.Contains(err.Error(), "acme") {
		t.Errorf("error should name the target project, got %q", err)
	}
}

func TestValidateDismissOptionsRejectsMalformedCanonical(t *testing.T) {
	if err := validateDismissOptions("acme/FB-2", "duplicate", "acme/1"); err == nil {
		t.Fatal("malformed canonical should be rejected locally")
	}
}

// not_planned ignores the canonical server-side, but a wrong value still points
// at a mistake worth surfacing before the request.
func TestValidateDismissOptionsValidatesCanonicalForNotPlanned(t *testing.T) {
	if err := validateDismissOptions("acme/FB-2", "not_planned", "other/FB-1"); err == nil {
		t.Fatal("cross-project canonical should be rejected regardless of reason")
	}
}

func TestValidateDismissOptionsRejectsMalformedTarget(t *testing.T) {
	if err := validateDismissOptions("FB-2", "duplicate", "acme/FB-1"); err == nil {
		t.Fatal("malformed target reference should be rejected locally")
	}
}

func TestDismissHasCanonicalFlag(t *testing.T) {
	if dismissCmd.Flags().Lookup("canonical") == nil {
		t.Fatal("dismiss should expose a --canonical flag")
	}
}
