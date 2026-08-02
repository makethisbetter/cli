package feedback

import "testing"

func TestValidateReadyOptionsRequiresSummary(t *testing.T) {
	if err := validateReadyOptions(""); err == nil {
		t.Fatal("ready should require a resolution summary")
	}
	if err := validateReadyOptions("   \t\n"); err == nil {
		t.Fatal("ready should reject a whitespace-only resolution summary")
	}
	if err := validateReadyOptions("Fixed Safari export."); err != nil {
		t.Fatalf("ready should accept a factual summary: %v", err)
	}
}

func TestReadyHasSummaryFlag(t *testing.T) {
	if readyCmd.Flags().Lookup("summary") == nil {
		t.Fatal("ready should expose a --summary flag")
	}
}

func TestHasFeedbackReferenceRequiresExactQualifiedReference(t *testing.T) {
	references := []string{"acme/FB-42", "other/FB-42"}
	if !hasFeedbackReference(references, "acme/FB-42") {
		t.Fatal("expected exact qualified reference to match")
	}
	if hasFeedbackReference(references, "acme/FB-4") {
		t.Fatal("partial feedback reference should not match")
	}
}
