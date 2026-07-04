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
		{"fb_123", "received"},
		{"fb_1", "in_progress"},
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

func TestPrintFeedbackResultJSON(t *testing.T) {
	var buf bytes.Buffer
	fb := &api.Feedback{ID: "fb_1", Status: "received", Priority: "high"}
	PrintFeedbackResult(&buf, fb, true, "ignored message")
	got := buf.String()
	if !strings.Contains(got, `"fb_1"`) {
		t.Errorf("expected JSON with fb_1, got %q", got)
	}
	if strings.Contains(got, "ignored message") {
		t.Error("message should not appear in JSON mode")
	}
}

func TestPrintFeedbackResultFormatted(t *testing.T) {
	var buf bytes.Buffer
	fb := &api.Feedback{ID: "fb_2", Status: "in_progress", Priority: "medium"}
	PrintFeedbackResult(&buf, fb, false, "Feedback fb_2 picked.")
	got := buf.String()
	if !strings.Contains(got, "Feedback fb_2 picked.") {
		t.Errorf("expected status message, got %q", got)
	}
	if !strings.Contains(got, "fb_2") {
		t.Errorf("expected detail output with fb_2, got %q", got)
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
