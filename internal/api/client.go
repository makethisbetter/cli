package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/makethisbetter/cli/internal/config"
)

const (
	requestTimeout                = 30 * time.Second
	projectHandleErrorMessage     = "project handle must be 4-63 lowercase letters, numbers, or internal hyphens; the third and fourth characters cannot both be hyphens"
	feedbackReferenceErrorMessage = "feedback reference must use {project-handle}/FB-{number}, for example acme/FB-42; the handle's third and fourth characters cannot both be hyphens"
)

var (
	projectHandlePattern     = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{2,61}[a-z0-9])$`)
	feedbackReferencePattern = regexp.MustCompile(`^([a-z0-9](?:[a-z0-9-]{2,61}[a-z0-9]))/FB-([1-9][0-9]*)$`)
)

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
	if params.ProjectHandle == "" {
		return nil, fmt.Errorf("project handle is required")
	}
	if err := validateProjectHandle(params.ProjectHandle); err != nil {
		return nil, err
	}

	q := c.withAccount(url.Values{})
	setIfPresent(q, "status", params.Status)
	setIfPresent(q, "label", params.Label)
	setIfPresent(q, "priority", params.Priority)

	var feedbacks []Feedback
	path := fmt.Sprintf("/projects/%s/feedbacks", url.PathEscape(params.ProjectHandle))
	if err := c.doJSON(ctx, "GET", path, nil, q, true, &feedbacks); err != nil {
		return nil, err
	}
	return feedbacks, nil
}

// GetFeedback returns a single feedback by project-scoped reference.
func (c *Client) GetFeedback(ctx context.Context, reference string) (*Feedback, error) {
	handle, number, err := ParseFeedbackReference(reference)
	if err != nil {
		return nil, err
	}

	q := c.withAccount(url.Values{})
	var fb Feedback
	path := fmt.Sprintf("/projects/%s/feedbacks/%s", url.PathEscape(handle), number)
	if err := c.doJSON(ctx, "GET", path, nil, q, true, &fb); err != nil {
		return nil, err
	}
	return &fb, nil
}

// UpdateFeedback applies params to the feedback with the given reference and returns
// the updated record.
func (c *Client) UpdateFeedback(ctx context.Context, reference string, params UpdateFeedbackParams) (*Feedback, error) {
	handle, number, err := ParseFeedbackReference(reference)
	if err != nil {
		return nil, err
	}

	q := c.withAccount(url.Values{})
	body := map[string]any{
		"feedback": params,
	}
	var fb Feedback
	path := fmt.Sprintf("/projects/%s/feedbacks/%s", url.PathEscape(handle), number)
	if err := c.doJSON(ctx, "PATCH", path, body, q, true, &fb); err != nil {
		return nil, err
	}
	return &fb, nil
}

// ResolveFeedback marks feedback shipped through the strict v2 resolution resource.
func (c *Client) ResolveFeedback(ctx context.Context, reference string, params ResolveFeedbackParams) (*Feedback, error) {
	handle, number, err := ParseFeedbackReference(reference)
	if err != nil {
		return nil, err
	}

	baseURL, err := c.baseURLForVersion("v2")
	if err != nil {
		return nil, err
	}
	q := c.withAccount(url.Values{})
	body := map[string]any{
		"feedback_resolution": params,
	}
	var fb Feedback
	path := fmt.Sprintf("/projects/%s/feedbacks/%s/resolution", url.PathEscape(handle), number)
	if err := c.doJSONAt(ctx, baseURL, "POST", path, body, q, true, &fb); err != nil {
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
	if err := validateProjectHandle(id); err != nil {
		return nil, err
	}

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
	if err := validateProjectHandle(params.Handle); err != nil {
		return nil, err
	}

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

// UpdateProject updates a project's mutable fields and returns the updated
// project. Only non-nil params are sent; the rest stay unchanged.
func (c *Client) UpdateProject(ctx context.Context, id string, params UpdateProjectParams) (*Project, error) {
	q := c.withAccount(url.Values{})
	body := map[string]any{
		"project": params,
	}
	var p Project
	path := fmt.Sprintf("/projects/%s", url.PathEscape(id))
	if err := c.doJSON(ctx, "PATCH", path, body, q, true, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, query url.Values, auth bool, out any) error {
	return c.doJSONAt(ctx, c.baseURL, method, path, body, query, auth, out)
}

func (c *Client) doJSONAt(ctx context.Context, baseURL, method, path string, body any, query url.Values, auth bool, out any) error {
	u, err := url.Parse(baseURL + path)
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

func (c *Client) baseURLForVersion(version string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid API URL: %w", err)
	}

	path := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(path, "/v1") {
		return "", fmt.Errorf("api_url must end in /v1 for feedback resolution; update ~/.makethisbetter/config.json")
	}
	u.Path = strings.TrimSuffix(path, "/v1") + "/" + version
	return strings.TrimSuffix(u.String(), "/"), nil
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

// ParseFeedbackReference splits a {project-handle}/FB-{number} reference into
// its handle and number. Commands use it to reject malformed references before
// a request is sent.
func ParseFeedbackReference(reference string) (string, string, error) {
	match := feedbackReferencePattern.FindStringSubmatch(reference)
	if match == nil || !validProjectHandle(match[1]) {
		return "", "", fmt.Errorf("%s", feedbackReferenceErrorMessage)
	}
	return match[1], match[2], nil
}

func validateProjectHandle(handle string) error {
	if !validProjectHandle(handle) {
		return fmt.Errorf("%s", projectHandleErrorMessage)
	}
	return nil
}

func validProjectHandle(handle string) bool {
	return projectHandlePattern.MatchString(handle) && handle[2:4] != "--"
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
	// A 401 is always about the stored token, so the login hint beats whatever
	// wording the server used.
	if status == http.StatusUnauthorized {
		return &APIError{
			StatusCode: status,
			Message:    "authentication failed, run `makethisbetter login` to re-authenticate",
		}
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return &APIError{StatusCode: status, Message: errResp.Error}
	}

	return &APIError{StatusCode: status, Message: emptyBodyErrorMessage(status)}
}

// emptyBodyErrorMessage covers the statuses the API can answer with no JSON
// body at all (`head :forbidden`, a missing record, a bare validation failure).
// A naked status number tells the caller nothing they can act on.
func emptyBodyErrorMessage(status int) string {
	switch status {
	case http.StatusForbidden:
		return "not authorized for this project or account, check `makethisbetter project list` and the account_id in ~/.makethisbetter/config.json"
	case http.StatusNotFound:
		return "not found, check the project handle and feedback number with `makethisbetter feedback list --project <handle>`"
	case http.StatusUnprocessableEntity:
		return "the API rejected the request as invalid, check the flag values and try again"
	default:
		return fmt.Sprintf("API request failed with status %d", status)
	}
}
