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
	ID                   string             `json:"id"`
	Reference            string             `json:"reference"`
	Number               int                `json:"number"`
	ProjectID            string             `json:"project_id"`
	ProjectHandle        string             `json:"project_handle"`
	Description          string             `json:"description"`
	AISummary            json.RawMessage    `json:"ai_structured_summary"`
	PageURL              *string            `json:"page_url"`
	UserAgent            *string            `json:"user_agent"`
	Browser              *string            `json:"browser"`
	OS                   *string            `json:"os"`
	ConsoleErrors        []json.RawMessage  `json:"console_errors"`
	TargetElement        json.RawMessage    `json:"target_element"`
	ReporterEmail        *string            `json:"reporter_email"`
	ReporterName         *string            `json:"reporter_name"`
	ReporterExternalID   *string            `json:"reporter_external_id"`
	Status               string             `json:"status"`
	Labels               []string           `json:"labels"`
	Priority             string             `json:"priority"`
	UpvotesCount         int                `json:"upvotes_count"`
	CreatedAt            string             `json:"created_at"`
	UpdatedAt            string             `json:"updated_at"`
	Recommendation       *string            `json:"recommendation"`
	CloseReason          *string            `json:"close_reason"`
	PRURL                *string            `json:"pr_url"`
	Assignee             *FeedbackAssignee  `json:"assignee"`
	ResolutionSummary    *string            `json:"resolution_summary"`
	CanonicalFeedback    *FeedbackReference `json:"canonical_feedback"`
	TerminalOutcomeAt    *string            `json:"terminal_outcome_at"`
	ClosingComment       *string            `json:"closing_comment"`
	ClosingCommentStatus string             `json:"closing_comment_status"`
	ScreenshotAttached   bool               `json:"screenshot_attached"`
	ScreenshotURL        *string            `json:"screenshot_url"`
	RecordingAttached    bool               `json:"recording_attached"`
	RecordingDuration    *int               `json:"recording_duration"`
	RecordingURL         *string            `json:"recording_url"`
	AIClarifyAvailable   bool               `json:"ai_clarify_available"`
	Markdown             string             `json:"markdown,omitempty"`

	// Evidence the reporter left behind. Anything the API sends but this struct
	// does not declare is silently dropped from --json, so these have to be
	// re-declared here field for field whenever the API grows.
	Annotations             []Annotation           `json:"annotations"`
	Breadcrumbs             []json.RawMessage      `json:"breadcrumbs"`
	AIClarificationMessages []ClarificationMessage `json:"ai_clarification_messages"`

	// Triage state. Without these a feedback whose triage failed looks exactly
	// like one that was never triaged.
	AITriageStatus        *string `json:"ai_triage_status"`
	AITriageError         *string `json:"ai_triage_error"`
	AIClarificationStatus *string `json:"ai_clarification_status"`

	// Reproducing a layout bug needs the viewport, not just the browser string.
	ScreenWidth      *int    `json:"screen_width"`
	ScreenHeight     *int    `json:"screen_height"`
	ReporterLanguage *string `json:"reporter_language"`
	ArchivedAt       *string `json:"archived_at"`
}

// Annotation is one point the reporter marked on the page, together with the
// element that point landed on.
type Annotation struct {
	X              float64         `json:"x"`
	Y              float64         `json:"y"`
	Type           string          `json:"type"`
	TargetName     string          `json:"targetName"`
	TargetText     string          `json:"targetText"`
	TargetSelector string          `json:"targetSelector"`
	TargetRect     *AnnotationRect `json:"targetRect"`
}

// AnnotationRect is the bounding box of an annotated element, in CSS pixels.
type AnnotationRect struct {
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Bottom float64 `json:"bottom"`
}

// ClarificationMessage is one turn of the AI's follow-up conversation with the
// reporter.
type ClarificationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ListFeedbacksParams holds the optional filters for listing feedback.
type ListFeedbacksParams struct {
	Status        string
	Label         string
	Priority      string
	ProjectHandle string
	AccountID     string
	Limit         int
}

// FeedbackResponseResult is returned after a one-way Reporter Notice is queued.
type FeedbackResponseResult struct {
	Feedback Feedback         `json:"feedback"`
	Delivery FeedbackDelivery `json:"delivery"`
}

// FeedbackResult wraps a Feedback returned by a dedicated operation resource.
type FeedbackResult struct {
	Feedback Feedback `json:"feedback"`
}

// FeedbackListResult wraps a dedicated Feedback collection.
type FeedbackListResult struct {
	Feedbacks []Feedback `json:"feedbacks"`
}

// FeedbackDelivery exposes delivery state without returning the response body.
type FeedbackDelivery struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RespondFeedbackParams holds the final user-provided Reporter Notice.
type RespondFeedbackParams struct {
	Subject *string `json:"subject,omitempty"`
	Body    string  `json:"body"`
}

// UpdateFeedbackParams holds the fields that can be changed on a feedback.
type UpdateFeedbackParams struct {
	Status              string `json:"status,omitempty"`
	Takeover            bool   `json:"takeover,omitempty"`
	CloseReason         string `json:"close_reason,omitempty"`
	PRURL               string `json:"pr_url,omitempty"`
	ResolutionSummary   string `json:"resolution_summary,omitempty"`
	CanonicalFeedbackID string `json:"canonical_feedback_id,omitempty"`
}

// FeedbackAssignee identifies the team member currently handling feedback.
type FeedbackAssignee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FeedbackReference identifies a related feedback in API responses.
type FeedbackReference struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
}

// ErrorResponse is the JSON error body returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Project is a single project returned by the API. APIKey and SigningSecret
// are only populated on show/create responses; SigningSecret is further
// restricted to authorized project managers and is nil when the caller lacks that capability.
type Project struct {
	ID                          string  `json:"id"`
	Handle                      string  `json:"handle"`
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
	AIContext                   *string `json:"ai_context,omitempty"`
}

// CreateProjectParams holds the fields accepted when creating a project.
type CreateProjectParams struct {
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	Domain    string `json:"domain,omitempty"`
	AIContext string `json:"ai_context,omitempty"`
}

// UpdateProjectParams holds the fields accepted when updating a project.
// Nil fields are omitted from the request and left unchanged by the server.
type UpdateProjectParams struct {
	Name      *string `json:"name,omitempty"`
	Domain    *string `json:"domain,omitempty"`
	AIContext *string `json:"ai_context,omitempty"`
}
