package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/makethisbetter/cli/internal/api"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestTruncateNewlines(t *testing.T) {
	got := truncate("line one\nline two", 48)
	if strings.Contains(got, "\n") {
		t.Errorf("expected newlines replaced, got %q", got)
	}
}

func TestPtrOr(t *testing.T) {
	s := "hello"
	if ptrOr(&s, "default") != "hello" {
		t.Error("expected pointer value")
	}
	if ptrOr(nil, "default") != "default" {
		t.Error("expected fallback")
	}
}

func TestFormatReporter(t *testing.T) {
	name := "Alice"
	email := "alice@example.com"

	tests := []struct {
		name string
		fb   api.Feedback
		want string
	}{
		{"both", api.Feedback{ReporterName: &name, ReporterEmail: &email}, "Alice <alice@example.com>"},
		{"name only", api.Feedback{ReporterName: &name}, "Alice"},
		{"email only", api.Feedback{ReporterEmail: &email}, "alice@example.com"},
		{"neither", api.Feedback{}, "-"},
	}
	for _, tt := range tests {
		got := formatReporter(&tt.fb)
		if got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	got := formatDate("2024-06-15T10:30:00Z")
	if got == "" || got == "2024-06-15T10:30:00Z" {
		t.Errorf("expected formatted date, got %q", got)
	}
	if !strings.Contains(got, "2024") {
		t.Errorf("expected year 2024 in output, got %q", got)
	}
}

func TestFormatDateInvalid(t *testing.T) {
	got := formatDate("not-a-date")
	if got != "not-a-date" {
		t.Errorf("expected passthrough for invalid date, got %q", got)
	}
}

func TestColumnWidths(t *testing.T) {
	rows := [][]string{
		{"ID", "Status"},
		{"FB-123", "received"},
		{"FB-1", "in_progress"},
	}
	widths := columnWidths(rows, []int{0, 0})
	if widths[0] != 6 {
		t.Errorf("col 0 width: got %d, want 6", widths[0])
	}
	if widths[1] != 11 {
		t.Errorf("col 1 width: got %d, want 11", widths[1])
	}
}

func TestColumnWidthsMaxCap(t *testing.T) {
	rows := [][]string{
		{"Description"},
		{"a very long description that exceeds max"},
	}
	widths := columnWidths(rows, []int{20})
	if widths[0] != 20 {
		t.Errorf("expected capped width 20, got %d", widths[0])
	}
}

func TestColorStatus(t *testing.T) {
	if ColorStatus("received") != "received" {
		t.Error("received should be unstyled")
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	PrintJSON(&buf, map[string]string{"key": "value"})
	got := buf.String()
	if !strings.Contains(got, `"key"`) || !strings.Contains(got, `"value"`) {
		t.Errorf("expected JSON output, got %q", got)
	}
}

func TestFeedbackTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	FeedbackTable(&buf, nil)
	got := buf.String()
	if !strings.Contains(got, "No feedback found") {
		t.Errorf("expected empty message, got %q", got)
	}
}

func TestFeedbackTableShowsLabels(t *testing.T) {
	var buf bytes.Buffer
	FeedbackTable(&buf, []api.Feedback{{Reference: "acme/FB-1", Labels: []string{"Bug", "Safari"}}})
	got := buf.String()
	if !strings.Contains(got, "Labels") || !strings.Contains(got, "Bug, Safari") {
		t.Errorf("expected labels in feedback table, got %q", got)
	}
}

func TestPrintFeedbackResultJSON(t *testing.T) {
	var buf bytes.Buffer
	fb := &api.Feedback{ID: "FB-1", Reference: "acme/FB-1", Status: "received", Priority: "high"}
	PrintFeedbackResult(&buf, fb, true, "ignored message")
	got := buf.String()
	if !strings.Contains(got, `"acme/FB-1"`) {
		t.Errorf("expected JSON with acme/FB-1, got %q", got)
	}
	if strings.Contains(got, "ignored message") {
		t.Error("message should not appear in JSON mode")
	}
}

func TestPrintFeedbackResultFormatted(t *testing.T) {
	var buf bytes.Buffer
	fb := &api.Feedback{ID: "FB-2", Reference: "acme/FB-2", Status: "in_progress", Priority: "medium"}
	PrintFeedbackResult(&buf, fb, false, "Feedback acme/FB-2 picked.")
	got := buf.String()
	if !strings.Contains(got, "Feedback acme/FB-2 picked.") {
		t.Errorf("expected status message, got %q", got)
	}
	if !strings.Contains(got, "acme/FB-2") {
		t.Errorf("expected detail output with acme/FB-2, got %q", got)
	}
}

func TestProjectTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	ProjectTable(&buf, nil)
	got := buf.String()
	if !strings.Contains(got, "No projects found") {
		t.Errorf("expected empty message, got %q", got)
	}
}

func TestProjectTableRows(t *testing.T) {
	var buf bytes.Buffer
	ProjectTable(&buf, []api.Project{
		{ID: "acme", Name: "Acme", FeedbacksCount: 3},
	})
	got := buf.String()
	if !strings.Contains(got, "acme") || !strings.Contains(got, "Acme") || !strings.Contains(got, "3") {
		t.Errorf("expected row with project fields, got %q", got)
	}
}

func TestPrintProjectResultJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &api.Project{ID: "acme", Name: "Acme"}
	PrintProjectResult(&buf, p, true, "ignored message")
	got := buf.String()
	if !strings.Contains(got, `"acme"`) {
		t.Errorf("expected JSON with acme, got %q", got)
	}
	if strings.Contains(got, "ignored message") {
		t.Error("message should not appear in JSON mode")
	}
}

func TestPrintProjectResultFormatted(t *testing.T) {
	var buf bytes.Buffer
	p := &api.Project{ID: "acme", Name: "Acme", APIKey: "mtb_proj_abc"}
	PrintProjectResult(&buf, p, false, "Project Acme created.")
	got := buf.String()
	if !strings.Contains(got, "Project Acme created.") {
		t.Errorf("expected status message, got %q", got)
	}
	if !strings.Contains(got, "mtb_proj_abc") {
		t.Errorf("expected api key in detail output, got %q", got)
	}
}

func TestProjectDetailSigningSecretAbsent(t *testing.T) {
	var buf bytes.Buffer
	ProjectDetail(&buf, &api.Project{ID: "acme", Name: "Acme"})
	got := buf.String()
	if !strings.Contains(got, "admin only") {
		t.Errorf("expected admin-only note when signing secret absent, got %q", got)
	}
}

func TestProjectDetailSigningSecretPresent(t *testing.T) {
	secret := "whsec_abc"
	var buf bytes.Buffer
	ProjectDetail(&buf, &api.Project{ID: "acme", Name: "Acme", SigningSecret: &secret})
	got := buf.String()
	if !strings.Contains(got, "whsec_abc") {
		t.Errorf("expected signing secret in output, got %q", got)
	}
}

func TestSuccessWriter(t *testing.T) {
	var buf bytes.Buffer
	Success(&buf, "it worked")
	got := buf.String()
	if !strings.Contains(got, "it worked") {
		t.Errorf("expected success message, got %q", got)
	}
}
