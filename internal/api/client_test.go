package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/makethisbetter/cli/internal/config"
)

func TestListFeedbacks(t *testing.T) {
	feedbacks := []Feedback{
		{ID: "fb_1", Status: "received", Priority: "high", Description: "Test feedback"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/feedbacks" {
			t.Errorf("expected /feedbacks, got %s", r.URL.Path)
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

	result, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 feedback, got %d", len(result))
	}
	if result[0].ID != "fb_1" {
		t.Errorf("expected ID fb_1, got %s", result[0].ID)
	}
}

func TestListFeedbacksWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("status") != "received" {
			t.Errorf("expected status=received, got %s", q.Get("status"))
		}
		if q.Get("feedback_type") != "bug" {
			t.Errorf("expected feedback_type=bug, got %s", q.Get("feedback_type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Feedback{})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{
		Status:       "received",
		FeedbackType: "bug",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFeedback(t *testing.T) {
	fb := Feedback{ID: "fb_42", Status: "in_progress", Description: "Detail"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feedbacks/fb_42" {
			t.Errorf("expected /feedbacks/fb_42, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fb)
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.GetFeedback(context.Background(), "fb_42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "fb_42" {
		t.Errorf("expected ID fb_42, got %s", result.ID)
	}
}

func TestUpdateFeedback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var body map[string]json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)

		var fb map[string]any
		json.Unmarshal(body["feedback"], &fb)
		if fb["status"] != "in_progress" {
			t.Errorf("expected status=in_progress, got %v", fb["status"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Feedback{ID: "fb_1", Status: "in_progress"})
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "tok", APIURL: server.URL})
	result, err := client.UpdateFeedback(context.Background(), "fb_1", UpdateFeedbackParams{Status: "in_progress"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "in_progress" {
		t.Errorf("expected status in_progress, got %s", result.Status)
	}
}

func TestAPIError401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	client := NewClient(&config.Config{Token: "bad", APIURL: server.URL})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{})
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
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "validation failed: status is invalid" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNoTokenError(t *testing.T) {
	client := NewClient(&config.Config{APIURL: "http://localhost"})
	_, err := client.ListFeedbacks(context.Background(), ListFeedbacksParams{})
	if err == nil {
		t.Fatal("expected error for missing token")
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
