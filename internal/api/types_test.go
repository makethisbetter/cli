package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFeedbackJSONRoundTrip(t *testing.T) {
	original := Feedback{
		ID:          "FB-42",
		Reference:   "acme/FB-42",
		ProjectID:   "acme",
		Description: "Button is broken",
		Status:      "received",
		Priority:    "high",
		CreatedAt:   "2024-06-01T12:00:00Z",
		UpdatedAt:   "2024-06-02T08:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Feedback
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.Priority != original.Priority {
		t.Errorf("Priority: got %q, want %q", decoded.Priority, original.Priority)
	}
	if decoded.Description != original.Description {
		t.Errorf("Description: got %q, want %q", decoded.Description, original.Description)
	}
}

func TestFeedbackJSONNullableFields(t *testing.T) {
	input := `{
			"id": "FB-1",
			"reference": "acme/FB-1",
			"number": 1,
			"project_id": "acme",
			"project_handle": "acme",
			"ai_clarify_available": true,
		"status": "received",
		"priority": "low",
		"description": "test",
		"page_url": "https://example.com",
		"reporter_email": "user@test.com",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z"
	}`

	var fb Feedback
	if err := json.Unmarshal([]byte(input), &fb); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if fb.PageURL == nil {
		t.Error("PageURL should not be nil")
	} else if *fb.PageURL != "https://example.com" {
		t.Errorf("PageURL: got %q, want %q", *fb.PageURL, "https://example.com")
	}

	if fb.ReporterEmail == nil {
		t.Error("ReporterEmail should not be nil")
	} else if *fb.ReporterEmail != "user@test.com" {
		t.Errorf("ReporterEmail: got %q, want %q", *fb.ReporterEmail, "user@test.com")
	}

	if fb.UserAgent != nil {
		t.Errorf("UserAgent should be nil when absent, got %v", fb.UserAgent)
	}
}

// The struct is what --json re-serializes from, so a field the struct does not
// declare disappears from --json without any error. This test pins the API
// payload against the struct so a future API field cannot go missing quietly.
func TestFeedbackEvidenceFieldsSurviveRoundTrip(t *testing.T) {
	input := `{
		"id": "FB-7",
		"annotations": [
			{
				"x": 120.5,
				"y": 340,
				"type": "point",
				"targetName": "Save button",
				"targetText": "Save changes",
				"targetSelector": "#form > button.primary",
				"targetRect": {"top": 10, "left": 20, "width": 100, "height": 40, "bottom": 50}
			}
		],
		"breadcrumbs": [
			{"type": "click", "at": "2024-01-01T00:00:00Z", "target": "nav a", "extra": {"nested": true}},
			{"type": "navigate", "to": "/settings"}
		],
		"ai_clarification_messages": [
			{"role": "assistant", "content": "Which button did you press?"},
			{"role": "user", "content": "The blue one"}
		],
		"ai_triage_status": "failed",
		"ai_triage_error": "Anthropic API timeout",
		"ai_clarification_status": "active",
		"screen_width": 1440,
		"screen_height": 900,
		"reporter_language": "zh-CN"
	}`

	var fb Feedback
	if err := json.Unmarshal([]byte(input), &fb); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(fb.Annotations) != 1 {
		t.Fatalf("Annotations: got %d, want 1", len(fb.Annotations))
	}
	a := fb.Annotations[0]
	if a.TargetName != "Save button" {
		t.Errorf("TargetName: got %q, want %q", a.TargetName, "Save button")
	}
	if a.TargetSelector != "#form > button.primary" {
		t.Errorf("TargetSelector: got %q, want %q", a.TargetSelector, "#form > button.primary")
	}
	if a.TargetText != "Save changes" {
		t.Errorf("TargetText: got %q, want %q", a.TargetText, "Save changes")
	}
	if a.Type != "point" {
		t.Errorf("Type: got %q, want %q", a.Type, "point")
	}
	if a.X != 120.5 || a.Y != 340 {
		t.Errorf("coordinates: got (%v, %v), want (120.5, 340)", a.X, a.Y)
	}
	if a.TargetRect == nil {
		t.Fatal("TargetRect should be populated")
	}
	if a.TargetRect.Bottom != 50 || a.TargetRect.Width != 100 {
		t.Errorf("TargetRect: got %+v, want bottom 50 / width 100", *a.TargetRect)
	}

	if len(fb.Breadcrumbs) != 2 {
		t.Fatalf("Breadcrumbs: got %d, want 2", len(fb.Breadcrumbs))
	}

	if len(fb.AIClarificationMessages) != 2 {
		t.Fatalf("AIClarificationMessages: got %d, want 2", len(fb.AIClarificationMessages))
	}
	if fb.AIClarificationMessages[1].Role != "user" || fb.AIClarificationMessages[1].Content != "The blue one" {
		t.Errorf("second clarification message: got %+v", fb.AIClarificationMessages[1])
	}

	if fb.AITriageStatus == nil || *fb.AITriageStatus != "failed" {
		t.Errorf("AITriageStatus: got %v, want failed", fb.AITriageStatus)
	}
	if fb.AITriageError == nil || *fb.AITriageError != "Anthropic API timeout" {
		t.Errorf("AITriageError: got %v, want the API timeout message", fb.AITriageError)
	}
	if fb.AIClarificationStatus == nil || *fb.AIClarificationStatus != "active" {
		t.Errorf("AIClarificationStatus: got %v, want active", fb.AIClarificationStatus)
	}
	if fb.ScreenWidth == nil || *fb.ScreenWidth != 1440 {
		t.Errorf("ScreenWidth: got %v, want 1440", fb.ScreenWidth)
	}
	if fb.ScreenHeight == nil || *fb.ScreenHeight != 900 {
		t.Errorf("ScreenHeight: got %v, want 900", fb.ScreenHeight)
	}
	if fb.ReporterLanguage == nil || *fb.ReporterLanguage != "zh-CN" {
		t.Errorf("ReporterLanguage: got %v, want zh-CN", fb.ReporterLanguage)
	}

	// Re-serialize the way --json does and confirm nothing was dropped.
	data, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	for _, key := range []string{
		"annotations", "breadcrumbs", "ai_clarification_messages",
		"ai_triage_status", "ai_triage_error", "ai_clarification_status",
		"screen_width", "screen_height", "reporter_language",
	} {
		if _, ok := out[key]; !ok {
			t.Errorf("--json output dropped key %q", key)
		}
	}

	// Breadcrumbs are loosely shaped; the nested key has to come back untouched.
	if !strings.Contains(string(out["breadcrumbs"]), `"nested":true`) {
		t.Errorf("breadcrumbs lost nested data: %s", string(out["breadcrumbs"]))
	}
}

func TestFeedbackEvidenceFieldsDefaultToEmpty(t *testing.T) {
	var fb Feedback
	if err := json.Unmarshal([]byte(`{"id":"FB-1"}`), &fb); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if fb.Annotations != nil {
		t.Errorf("Annotations should be nil when absent, got %v", fb.Annotations)
	}
	if fb.AITriageStatus != nil {
		t.Errorf("AITriageStatus should be nil when absent, got %v", *fb.AITriageStatus)
	}
	if fb.ScreenWidth != nil {
		t.Errorf("ScreenWidth should be nil when absent, got %v", *fb.ScreenWidth)
	}
}

func TestUpdateFeedbackParamsOmitempty(t *testing.T) {
	tests := []struct {
		name       string
		params     UpdateFeedbackParams
		wantKey    string
		wantAbsent string
	}{
		{
			name:       "status only omits close reason",
			params:     UpdateFeedbackParams{Status: "in_progress"},
			wantKey:    "status",
			wantAbsent: "close_reason",
		},
		{
			name:       "close reason only omits status",
			params:     UpdateFeedbackParams{CloseReason: "shipped"},
			wantKey:    "close_reason",
			wantAbsent: "status",
		},
		{
			name:       "empty params omits both",
			params:     UpdateFeedbackParams{},
			wantKey:    "",
			wantAbsent: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.params)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var m map[string]json.RawMessage
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("Unmarshal to map failed: %v", err)
			}

			if tt.wantKey != "" {
				if _, ok := m[tt.wantKey]; !ok {
					t.Errorf("expected key %q in JSON output", tt.wantKey)
				}
			}

			if tt.wantAbsent != "" {
				if _, ok := m[tt.wantAbsent]; ok {
					t.Errorf("key %q should be omitted from JSON output", tt.wantAbsent)
				}
			}
		})
	}
}

func TestUpdateFeedbackParamsEmptyIsEmptyJSON(t *testing.T) {
	data, err := json.Marshal(UpdateFeedbackParams{})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("empty params should marshal to {}, got %s", string(data))
	}
}

func TestProjectJSONNullableFields(t *testing.T) {
	input := `{
		"id": "acme",
		"name": "Acme",
		"domain": "acme.com",
		"feedback_visibility": "public",
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z",
		"feedbacks_count": 3
	}`

	var p Project
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if p.Domain == nil || *p.Domain != "acme.com" {
		t.Errorf("Domain: got %v, want acme.com", p.Domain)
	}
	if p.SigningSecret != nil {
		t.Errorf("SigningSecret should be nil when absent, got %v", *p.SigningSecret)
	}
	if p.BoardURL != nil {
		t.Errorf("BoardURL should be nil when absent, got %v", *p.BoardURL)
	}
	if p.APIKey != "" {
		t.Errorf("APIKey should be empty when absent, got %q", p.APIKey)
	}
}

func TestProjectJSONWithAdminFields(t *testing.T) {
	input := `{
		"id": "acme",
		"name": "Acme",
		"api_key": "mtb_proj_abc",
		"board_url": "https://acme.makethisbetter.dev",
		"signing_secret": "whsec_abc"
	}`

	var p Project
	if err := json.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if p.APIKey != "mtb_proj_abc" {
		t.Errorf("APIKey: got %q, want mtb_proj_abc", p.APIKey)
	}
	if p.BoardURL == nil || *p.BoardURL != "https://acme.makethisbetter.dev" {
		t.Errorf("BoardURL: got %v, want https://acme.makethisbetter.dev", p.BoardURL)
	}
	if p.SigningSecret == nil || *p.SigningSecret != "whsec_abc" {
		t.Errorf("SigningSecret: got %v, want whsec_abc", p.SigningSecret)
	}
}

func TestCreateProjectParamsOmitempty(t *testing.T) {
	data, err := json.Marshal(CreateProjectParams{Name: "New", Handle: "new-project"})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if _, ok := m["name"]; !ok {
		t.Error("expected key \"name\" in JSON output")
	}
	if _, ok := m["handle"]; !ok {
		t.Error("expected key \"handle\" in JSON output")
	}
	if _, ok := m["domain"]; ok {
		t.Error("domain should be omitted when empty")
	}
}

func TestFeedbackJSONFieldNames(t *testing.T) {
	fb := Feedback{
		ID:                 "FB-1",
		Reference:          "acme/FB-1",
		ScreenshotAttached: true,
		UpvotesCount:       5,
	}
	data, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedKeys := []string{
		"id", "reference", "number", "project_id", "project_handle",
		"ai_clarify_available", "screenshot_attached", "upvotes_count", "status", "priority",
	}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}
