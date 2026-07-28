package feedback

import "testing"

func TestValidateResolveOptionsRequiresSummary(t *testing.T) {
	if err := validateResolveOptions(""); err == nil {
		t.Fatal("resolve should require a resolution summary")
	}
	if err := validateResolveOptions("Fixed Safari export."); err != nil {
		t.Fatalf("resolve should accept a factual summary: %v", err)
	}
}

func TestResolveHasSummaryFlag(t *testing.T) {
	if resolveCmd.Flags().Lookup("summary") == nil {
		t.Fatal("resolve should expose a --summary flag")
	}
}
