package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/makethisbetter/cli/internal/config"
)

func TestFeedbackRespondReadsExactBodyFileAndQueuesResponse(t *testing.T) {
	secretBody := "First line\r\n\r\n**Second line.**\r\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/acme/feedbacks/42/response" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("account_id") != "acc_1" {
			t.Errorf("account_id = %q", r.URL.Query().Get("account_id"))
		}
		var request struct {
			FeedbackResponse struct {
				Body string `json:"body"`
			} `json:"feedback_response"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.FeedbackResponse.Body != "First line\n\n**Second line.**" {
			t.Errorf("body = %q", request.FeedbackResponse.Body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"feedback": map[string]any{"id": "FB-42", "reference": "acme/FB-42", "status": "closed", "close_reason": "responded"},
			"delivery": map[string]any{"id": "delivery_1", "status": "pending", "created_at": "2026-08-03T12:00:00Z"},
		})
	}))
	t.Cleanup(server.Close)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	if err := config.SaveTo(&config.Config{Token: "token", APIURL: server.URL + "/api/v1", AccountID: "acc_1"}, configPath); err != nil {
		t.Fatal(err)
	}
	bodyPath := filepath.Join(tempDir, "response.md")
	if err := os.WriteFile(bodyPath, []byte(secretBody), 0o600); err != nil {
		t.Fatal(err)
	}

	output, err := runLoginCLIResult(t, configPath, "feedback", "respond", "acme/FB-42", "--body-file", bodyPath)
	if err != nil {
		t.Fatalf("respond failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Response queued and feedback acme/FB-42 closed.") {
		t.Fatalf("output = %q", output)
	}
	if strings.Contains(output, "First line") || strings.Contains(output, "Second line") {
		t.Fatalf("output leaked response body: %q", output)
	}
}

func TestFeedbackArchiveAndRestoreUseTheSingleFeedbackResource(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/acme/feedbacks/42/archive" {
			t.Errorf("path = %s", r.URL.Path)
		}
		methods = append(methods, r.Method)
		archivedAt := any("2026-08-03T12:00:00Z")
		if r.Method == http.MethodDelete {
			archivedAt = nil
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"feedback": map[string]any{
				"id": "FB-42", "reference": "acme/FB-42", "status": "received", "archived_at": archivedAt,
			},
		})
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveTo(&config.Config{Token: "token", APIURL: server.URL + "/api/v1", AccountID: "acc_1"}, configPath); err != nil {
		t.Fatal(err)
	}

	archiveOutput, err := runLoginCLIResult(t, configPath, "feedback", "archive", "acme/FB-42")
	if err != nil {
		t.Fatalf("archive failed: %v\n%s", err, archiveOutput)
	}
	if !strings.Contains(archiveOutput, "Feedback acme/FB-42 is archived.") {
		t.Fatalf("archive output = %q", archiveOutput)
	}

	restoreOutput, err := runLoginCLIResult(t, configPath, "feedback", "restore", "acme/FB-42")
	if err != nil {
		t.Fatalf("restore failed: %v\n%s", err, restoreOutput)
	}
	if !strings.Contains(restoreOutput, "Feedback acme/FB-42 is active (not archived).") {
		t.Fatalf("restore output = %q", restoreOutput)
	}

	if len(methods) != 2 || methods[0] != http.MethodPost || methods[1] != http.MethodDelete {
		t.Fatalf("methods = %v, want [POST DELETE]", methods)
	}
}

func TestFeedbackListArchivedUsesTheArchivedCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/acme/archived_feedbacks" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("label") != "Bug" || r.URL.Query().Get("priority") != "high" || r.URL.Query().Get("limit") != "1" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"feedbacks": []map[string]any{{
				"id": "FB-42", "reference": "acme/FB-42", "status": "received", "archived_at": "2026-08-03T12:00:00Z",
			}},
		})
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveTo(&config.Config{Token: "token", APIURL: server.URL + "/api/v1"}, configPath); err != nil {
		t.Fatal(err)
	}

	output, err := runLoginCLIResult(t, configPath,
		"feedback", "list", "--project", "acme", "--archived", "--label", "Bug", "--priority", "high", "--limit", "1", "--json")
	if err != nil {
		t.Fatalf("archived list failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, `"archived_at": "2026-08-03T12:00:00Z"`) {
		t.Fatalf("output = %q", output)
	}
}

func TestFeedbackShowFallsBackToArchivedDetail(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/projects/acme/feedbacks/42" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"feedback": map[string]any{
				"id": "FB-42", "reference": "acme/FB-42", "status": "received",
				"archived_at": "2026-08-03T12:00:00Z", "markdown": "# acme/FB-42 — received\n\n- Archived at: 2026-08-03T12:00:00Z\n",
			},
		})
	}))
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveTo(&config.Config{Token: "token", APIURL: server.URL + "/api/v1"}, configPath); err != nil {
		t.Fatal(err)
	}

	output, err := runLoginCLIResult(t, configPath, "feedback", "show", "acme/FB-42", "--md")
	if err != nil {
		t.Fatalf("archived show failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Archived at: 2026-08-03T12:00:00Z") {
		t.Fatalf("output = %q", output)
	}
	wantPaths := []string{"/api/v1/projects/acme/feedbacks/42", "/api/v1/projects/acme/archived_feedbacks/42"}
	if len(paths) != 2 || paths[0] != wantPaths[0] || paths[1] != wantPaths[1] {
		t.Fatalf("paths = %v, want %v", paths, wantPaths)
	}
}
