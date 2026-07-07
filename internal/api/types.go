package api

import "encoding/json"

// RegistrationResponse is returned when requesting an OTP registration.
type RegistrationResponse struct {
	RegistrationToken string `json:"registration_token"`
	ExpiresIn         int    `json:"expires_in"`
}

// RegistrationVerifyResponse is returned after a successful OTP verification.
type RegistrationVerifyResponse struct {
	User     VerifyUser     `json:"user"`
	Account  VerifyAccount  `json:"account"`
	APIToken VerifyAPIToken `json:"api_token"`
}

// VerifyUser is the user record returned by registration verification.
type VerifyUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// VerifyAccount is the account record returned by registration verification.
type VerifyAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VerifyAPIToken is the API token issued by registration verification.
type VerifyAPIToken struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// Feedback is a single feedback item returned by the API.
type Feedback struct {
	ID                 string            `json:"id"`
	ProjectID          string            `json:"project_id"`
	Description        string            `json:"description"`
	AISummary          json.RawMessage   `json:"ai_structured_summary"`
	PageURL            *string           `json:"page_url"`
	UserAgent          *string           `json:"user_agent"`
	Browser            *string           `json:"browser"`
	OS                 *string           `json:"os"`
	ConsoleErrors      []json.RawMessage `json:"console_errors"`
	TargetElement      json.RawMessage   `json:"target_element"`
	ReporterEmail      *string           `json:"reporter_email"`
	ReporterName       *string           `json:"reporter_name"`
	ReporterExternalID *string           `json:"reporter_external_id"`
	Status             string            `json:"status"`
	Labels             json.RawMessage   `json:"labels"`
	Priority           string            `json:"priority"`
	UpvotesCount       int               `json:"upvotes_count"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
	FeedbackType       *string           `json:"feedback_type"`
	Recommendation     *string           `json:"recommendation"`
	CloseReason        *string           `json:"close_reason"`
	CustomLabels       []json.RawMessage `json:"custom_labels"`
	ScreenshotAttached bool              `json:"screenshot_attached"`
	RecordingAttached  bool              `json:"recording_attached"`
	RecordingDuration  *int              `json:"recording_duration"`
	RecordingURL       *string           `json:"recording_url"`
	Markdown           string            `json:"markdown,omitempty"`
}

// ListFeedbacksParams holds the optional filters for listing feedback.
type ListFeedbacksParams struct {
	Status       string
	FeedbackType string
	Priority     string
	ProjectID    string
	AccountID    string
}

// UpdateFeedbackParams holds the fields that can be changed on a feedback.
type UpdateFeedbackParams struct {
	Status string         `json:"status,omitempty"`
	Labels map[string]any `json:"labels,omitempty"`
}

// ErrorResponse is the JSON error body returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Project is a single project returned by the API. APIKey and SigningSecret
// are only populated on show/create responses; SigningSecret is further
// restricted to account admins and is nil when the caller lacks that role.
type Project struct {
	ID                          string  `json:"id"`
	Name                        string  `json:"name"`
	Domain                      *string `json:"domain"`
	FeedbackVisibility          string  `json:"feedback_visibility"`
	CreatedAt                   string  `json:"created_at"`
	UpdatedAt                   string  `json:"updated_at"`
	FeedbacksCount              int     `json:"feedbacks_count"`
	APIKey                      string  `json:"api_key,omitempty"`
	BoardURL                    *string `json:"board_url,omitempty"`
	EnforceIdentityVerification bool    `json:"enforce_identity_verification"`
	SigningSecret               *string `json:"signing_secret,omitempty"`
}

// CreateProjectParams holds the fields accepted when creating a project.
type CreateProjectParams struct {
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
}
