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
