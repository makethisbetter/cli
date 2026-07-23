package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/makethisbetter/cli/internal/config"
)

func TestListFeedbacks(t *testing.T) {
	feedbacks := []Feedback{
		{ID: "FB-1", Reference: "acme/FB-1", ProjectID: "acme", Status: "received", Priority: "high", Description: "Test feedback"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/projects/acme/feedbacks" {
			t.Errorf("expected /projects/acme/feedbacks, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or wrong auth header: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("account_id") != "acc_1" {
			t.Errorf("expected account_id=acc_1, got %s", r.URL.Query().Get("account_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(feedbacks)
	}))
	defer server.Close()

	client := NewClient(&config.Config{
		Token:     "test-token",
		APIURL:    server.URL,
		AccountID: "acc_1",
	})

	result, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{ProjectHandle: "acme"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 feedback, got %d", len(result))
	}
	if result[0].Reference != "acme/FB-1" {
		t.Errorf("expected reference acme/FB-1, got %s", result[0].Reference)
	}
}

func TestListFeedbacksWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("status") != "received" {
			t.Errorf("expected status=received, got %s", q.Get("status"))
		}
		if q.Get("label") != "Safari" {
			t.Errorf("expected label=Safari, got %s", q.Get("label"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Feedback{})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{
		ProjectHandle: "acme",
		Status:        "received",
		Label:         "Safari",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFeedback(t *testing.T) {
	fb := Feedback{ID: "FB-42", Reference: "acme/FB-42", ProjectID: "acme", Status: "in_progress", Description: "Detail"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/acme/feedbacks/42" {
			t.Errorf("expected /projects/acme/feedbacks/42, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fb)
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.GetFeedback(context.Background(), "acme/FB-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Reference != "acme/FB-42" {
		t.Errorf("expected reference acme/FB-42, got %s", result.Reference)
	}
}

func TestUpdateFeedback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/projects/acme/feedbacks/1" {
			t.Errorf("expected /projects/acme/feedbacks/1, got %s", r.URL.Path)
		}

		var body map[string]json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)

		var fb map[string]any
		json.Unmarshal(body["feedback"], &fb)
		if fb["status"] != "in_progress" {
			t.Errorf("expected status=in_progress, got %v", fb["status"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Feedback{ID: "FB-1", Reference: "acme/FB-1", ProjectID: "acme", Status: "in_progress"})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.UpdateFeedback(context.Background(), "acme/FB-1", UpdateFeedbackParams{Status: "in_progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "in_progress" {
		t.Errorf("expected status in_progress, got %s", result.Status)
	}
}

func TestGetFeedbackRejectsInvalidReference(t *testing.T) {
	client := NewClient(&config.Config{Token: "tok", APIURL: "http://localhost"})

	for _, reference := range []string{"FB-1", "abc/FB-1", "ab--cd/FB-1", strings.Repeat("a", 64) + "/FB-1"} {
		_, err := client.GetFeedback(context.Background(), reference)

		if err == nil || err.Error() != feedbackReferenceErrorMessage {
			t.Fatalf("expected invalid reference error for %q, got %v", reference, err)
		}
	}
}

func TestCreateProjectRejectsInvalidHandleBeforeRequest(t *testing.T) {
	requested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	for _, handle := range []string{"abc", "ab--cd", strings.Repeat("a", 64)} {
		_, err := client.CreateProject(context.Background(), CreateProjectParams{Name: "Invalid", Handle: handle})

		if err == nil || !strings.Contains(err.Error(), "project handle must") {
			t.Fatalf("expected invalid handle error for %q, got %v", handle, err)
		}
	}
	_, err := client.CreateProject(context.Background(), CreateProjectParams{Name: "Invalid", Handle: "ab--cd"})
	if err == nil || !strings.Contains(err.Error(), "third and fourth characters cannot both be hyphens") {
		t.Fatalf("expected R-LDH explanation, got %v", err)
	}
	if requested {
		t.Fatal("invalid handles must not reach the API")
	}
}

func TestAPIError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "bad", APIURL: server.URL})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{ProjectHandle: "acme"})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	expected := "authentication failed, run `makethisbetter login` to re-authenticate"
	if apiErr.Message != expected {
		t.Errorf("got %q, want %q", apiErr.Message, expected)
	}
}

func TestAPIErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"error":"validation failed: status is invalid"}`))
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{ProjectHandle: "acme"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "validation failed: status is invalid" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNoTokenError(t *testing.T) {
	client := NewClient(&config.Config{APIURL: "http://localhost"})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{ProjectHandle: "acme"})
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestListProjects(t *testing.T) {
	projects := []Project{
		{ID: "acme", Name: "Acme", FeedbacksCount: 3},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Errorf("expected /projects, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or wrong auth header: %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "test-token", APIURL: server.URL})

	result, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 project, got %d", len(result))
	}
	if result[0].ID != "acme" {
		t.Errorf("expected ID acme, got %s", result[0].ID)
	}
}

func TestGetProject(t *testing.T) {
	secret := "whsec_abc"
	project := Project{ID: "acme", Name: "Acme", APIKey: "mtb_proj_abc", SigningSecret: &secret}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/acme" {
			t.Errorf("expected /projects/acme, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.GetProject(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "acme" {
		t.Errorf("expected ID acme, got %s", result.ID)
	}
	if result.SigningSecret == nil || *result.SigningSecret != "whsec_abc" {
		t.Errorf("expected signing secret whsec_abc, got %v", result.SigningSecret)
	}
}

func TestGetProjectWithoutSigningSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"acme","name":"Acme","api_key":"mtb_proj_abc"}`))
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.GetProject(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SigningSecret != nil {
		t.Errorf("expected nil signing secret when absent, got %v", *result.SigningSecret)
	}
}

func TestCreateProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Errorf("expected /projects, got %s", r.URL.Path)
		}

		var body map[string]json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)

		var p map[string]any
		json.Unmarshal(body["project"], &p)
		if p["name"] != "New" {
			t.Errorf("expected name=New, got %v", p["name"])
		}
		if p["handle"] != "new-project" {
			t.Errorf("expected handle=new-project, got %v", p["handle"])
		}
		if p["domain"] != "example.com" {
			t.Errorf("expected domain=example.com, got %v", p["domain"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(Project{ID: "new-project", Handle: "new-project", Name: "New", APIKey: "mtb_proj_abc"})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.CreateProject(context.Background(), CreateProjectParams{Name: "New", Handle: "new-project", Domain: "example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "New" {
		t.Errorf("expected name New, got %s", result.Name)
	}
	if result.APIKey != "mtb_proj_abc" {
		t.Errorf("expected api key mtb_proj_abc, got %s", result.APIKey)
	}
}

func TestCreateProjectAccepts63CharacterHandle(t *testing.T) {
	handle := strings.Repeat("a", 63)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request: %v", err)
		}

		var project CreateProjectParams
		if err := json.Unmarshal(body["project"], &project); err != nil {
			t.Fatalf("decoding project: %v", err)
		}
		if project.Handle != handle {
			t.Errorf("expected 63-character handle, got %q", project.Handle)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Project{ID: handle, Handle: handle, Name: "Longest"})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	project, err := client.CreateProject(context.Background(), CreateProjectParams{Name: "Longest", Handle: handle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Handle != handle {
		t.Errorf("expected 63-character handle in response, got %q", project.Handle)
	}
}

func TestCreateProjectError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"not an account admin"}`))
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	_, err := client.CreateProject(context.Background(), CreateProjectParams{Name: "New", Handle: "new-project"})
	if err == nil {
		t.Fatal("expected error for 403")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

func TestRegistration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/agent_registration":
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Authorization") != "" {
				t.Error("registration should not send auth header")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(RegistrationResponse{
				RegistrationToken: "reg_123",
				ExpiresIn:         300,
			})
		case "/agent_registration/verify":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(RegistrationVerifyResponse{
				User:     VerifyUser{ID: "u_1", Email: "test@example.com"},
				Account:  VerifyAccount{ID: "acc_1", Name: "Test"},
				APIToken: VerifyAPIToken{Token: "tok_abc", Name: "CLI"},
			})
		}
	}))
	defer server.Close()

	client := NewUnauthClient(server.URL)

	reg, err := client.RequestRegistration(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("RequestRegistration failed: %v", err)
	}
	if reg.RegistrationToken != "reg_123" {
		t.Errorf("expected reg_123, got %s", reg.RegistrationToken)
	}

	verify, err := client.VerifyRegistration(context.Background(), reg.RegistrationToken, "123456")
	if err != nil {
		t.Fatalf("VerifyRegistration failed: %v", err)
	}
	if verify.APIToken.Token != "tok_abc" {
		t.Errorf("expected tok_abc, got %s", verify.APIToken.Token)
	}
	if verify.User.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", verify.User.Email)
	}
}
