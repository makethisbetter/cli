package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/makethisbetter/cli/internal/config"
)

const requestTimeout = 30 * time.Second

// Client talks to the Make This Better HTTP API.
type Client struct {
	baseURL    string
	token      string
	accountID  string
	httpClient *http.Client
}

// NewClient returns a Client authenticated with the token and account from cfg.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:   cfg.APIURL,
		token:     cfg.Token,
		accountID: cfg.AccountID,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// NewUnauthClient returns a Client for endpoints that do not require a token,
// such as registration and OTP verification.
func NewUnauthClient(apiURL string) *Client {
	return &Client{
		baseURL: config.NormalizeURL(apiURL),
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// RequestRegistration asks the API to send an OTP to the given email and
// returns the registration token used to verify it.
func (c *Client) RequestRegistration(ctx context.Context, email string) (*RegistrationResponse, error) {
	body := map[string]string{"email": email}
	var resp RegistrationResponse
	if err := c.doJSON(ctx, "POST", "/agent_registration", body, nil, false, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VerifyRegistration exchanges a registration token and OTP for an API token.
func (c *Client) VerifyRegistration(ctx context.Context, regToken, otp string) (*RegistrationVerifyResponse, error) {
	body := map[string]string{
		"registration_token": regToken,
		"otp":                otp,
	}
	var resp RegistrationVerifyResponse
	if err := c.doJSON(ctx, "POST", "/agent_registration/verify", body, nil, false, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListFeedbacks returns feedback matching the given filters.
func (c *Client) ListFeedbacks(ctx context.Context, params ListFeedbacksParams) ([]Feedback, error) {
	q := c.withAccount(url.Values{})
	setIfPresent(q, "status", params.Status)
	setIfPresent(q, "feedback_type", params.FeedbackType)
	setIfPresent(q, "priority", params.Priority)
	setIfPresent(q, "project_id", params.ProjectID)

	var feedbacks []Feedback
	if err := c.doJSON(ctx, "GET", "/feedbacks", nil, q, true, &feedbacks); err != nil {
		return nil, err
	}
	return feedbacks, nil
}

// GetFeedback returns a single feedback by id.
func (c *Client) GetFeedback(ctx context.Context, id string) (*Feedback, error) {
	q := c.withAccount(url.Values{})
	var fb Feedback
	path := fmt.Sprintf("/feedbacks/%s", url.PathEscape(id))
	if err := c.doJSON(ctx, "GET", path, nil, q, true, &fb); err != nil {
		return nil, err
	}
	return &fb, nil
}

// UpdateFeedback applies params to the feedback with the given id and returns
// the updated record.
func (c *Client) UpdateFeedback(ctx context.Context, id string, params UpdateFeedbackParams) (*Feedback, error) {
	q := c.withAccount(url.Values{})
	body := map[string]any{
		"feedback": params,
	}
	var fb Feedback
	path := fmt.Sprintf("/feedbacks/%s", url.PathEscape(id))
	if err := c.doJSON(ctx, "PATCH", path, body, q, true, &fb); err != nil {
		return nil, err
	}
	return &fb, nil
}

// ListProjects returns the account's projects.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	q := c.withAccount(url.Values{})
	var projects []Project
	if err := c.doJSON(ctx, "GET", "/projects", nil, q, true, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject returns a single project by id.
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	q := c.withAccount(url.Values{})
	var p Project
	path := fmt.Sprintf("/projects/%s", url.PathEscape(id))
	if err := c.doJSON(ctx, "GET", path, nil, q, true, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProject creates a new project and returns it.
func (c *Client) CreateProject(ctx context.Context, params CreateProjectParams) (*Project, error) {
	q := c.withAccount(url.Values{})
	body := map[string]any{
		"project": params,
	}
	var p Project
	if err := c.doJSON(ctx, "POST", "/projects", body, q, true, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, query url.Values, auth bool, out any) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		if c.token == "" {
			return config.ErrNotLoggedIn
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}

func (c *Client) withAccount(q url.Values) url.Values {
	if c.accountID != "" {
		q.Set("account_id", c.accountID)
	}
	return q
}

func setIfPresent(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

// APIError is returned when the API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Message    string
}

// Error returns the human-readable API error message.
func (e *APIError) Error() string {
	return e.Message
}

func parseAPIError(status int, body []byte) error {
	if status == 401 {
		return &APIError{
			StatusCode: status,
			Message:    "authentication failed, run `makethisbetter login` to re-authenticate",
		}
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return &APIError{StatusCode: status, Message: errResp.Error}
	}

	return &APIError{
		StatusCode: status,
		Message:    fmt.Sprintf("API request failed with status %d", status),
	}
}
