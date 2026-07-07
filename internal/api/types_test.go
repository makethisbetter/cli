package api

import (
	"encoding/json"
	"testing"
)

func TestFeedbackJSONRoundTrip(t *testing.T) {
	original := Feedback{
		ID:          "fb_42",
		ProjectID:   "proj_1",
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
		"id": "fb_1",
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

func TestUpdateFeedbackParamsOmitempty(t *testing.T) {
	tests := []struct {
		name       string
		params     UpdateFeedbackParams
		wantKey    string
		wantAbsent string
	}{
		{
			name:       "status only omits labels",
			params:     UpdateFeedbackParams{Status: "in_progress"},
			wantKey:    "status",
			wantAbsent: "labels",
		},
		{
			name:       "labels only omits status",
			params:     UpdateFeedbackParams{Labels: map[string]any{"close_reason": "shipped"}},
			wantKey:    "labels",
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
		"id": "project_1",
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
		"id": "project_1",
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
	data, err := json.Marshal(CreateProjectParams{Name: "New"})
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
	if _, ok := m["domain"]; ok {
		t.Error("domain should be omitted when empty")
	}
}

func TestFeedbackJSONFieldNames(t *testing.T) {
	fb := Feedback{
		ID:                 "fb_1",
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

	expectedKeys := []string{"id", "screenshot_attached", "upvotes_count", "project_id", "status", "priority"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}
