package feedback

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/spf13/cobra"
)

type fakeReleaseClient struct {
	feedbacks map[string]*api.Feedback
	failures  map[string]error
	requested []string
	released  []feedbackTrailer
}

func (client *fakeReleaseClient) GetFeedback(_ context.Context, reference string) (*api.Feedback, error) {
	client.requested = append(client.requested, reference)
	feedback, ok := client.feedbacks[reference]
	if !ok {
		return nil, errors.New("not found")
	}
	return feedback, nil
}

func (client *fakeReleaseClient) ReleaseFeedback(_ context.Context, reference, committedAt string) (*api.Feedback, error) {
	if err := client.failures[reference]; err != nil {
		return nil, err
	}
	timestamp, _ := time.Parse(time.RFC3339, committedAt)
	client.released = append(client.released, feedbackTrailer{Reference: reference, CommittedAt: timestamp})
	return &api.Feedback{Reference: reference, Status: "closed"}, nil
}

func TestReleaseFeedbackReferencesOnlyReleasesPendingFeedback(t *testing.T) {
	client := &fakeReleaseClient{
		feedbacks: map[string]*api.Feedback{
			"acme/FB-1": {Reference: "acme/FB-1", Status: "pending_release"},
			"acme/FB-2": {Reference: "acme/FB-2", Status: "received"},
			"acme/FB-3": {Reference: "acme/FB-3", Status: "pending_release"},
		},
		failures: map[string]error{"acme/FB-3": errors.New("unavailable")},
	}

	committedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	trailers := []feedbackTrailer{
		{Reference: "acme/FB-1", CommittedAt: committedAt},
		{Reference: "acme/FB-2", CommittedAt: committedAt},
		{Reference: "acme/FB-3", CommittedAt: committedAt},
	}
	result := releaseFeedbackReferences(context.Background(), client, trailers, "acme")

	if !reflect.DeepEqual(result.Released, []string{"acme/FB-1"}) {
		t.Fatalf("released %v", result.Released)
	}
	if !reflect.DeepEqual(result.Skipped, []releaseSkip{{Reference: "acme/FB-2", Reason: "not_pending_release"}}) {
		t.Fatalf("skipped %v", result.Skipped)
	}
	if len(result.Failures) != 1 || result.Failures[0].Reference != "acme/FB-3" {
		t.Fatalf("failures %v", result.Failures)
	}
	if !reflect.DeepEqual(client.released, []feedbackTrailer{{Reference: "acme/FB-1", CommittedAt: committedAt}}) {
		t.Fatalf("release calls %v", client.released)
	}
}

func TestReleaseFeedbackReferencesSkipsStaleTrailerConflict(t *testing.T) {
	client := &fakeReleaseClient{
		feedbacks: map[string]*api.Feedback{
			"acme/FB-1": {Reference: "acme/FB-1", Status: "pending_release"},
		},
		failures: map[string]error{
			"acme/FB-1": &api.APIError{StatusCode: 409, Message: "stale trailer"},
		},
	}

	result := releaseFeedbackReferences(context.Background(), client, []feedbackTrailer{{
		Reference: "acme/FB-1", CommittedAt: time.Now(),
	}}, "acme")

	want := []releaseSkip{{Reference: "acme/FB-1", Reason: "stale_trailer_before_reopen"}}
	if !reflect.DeepEqual(result.Skipped, want) || len(result.Failures) != 0 {
		t.Fatalf("unexpected release result: %#v", result)
	}
}

func TestReleaseCommandRequiresThrough(t *testing.T) {
	if releaseCmd.Flags().Lookup("through") == nil {
		t.Fatal("release should expose a --through flag")
	}
}

func TestReleaseFeedbackReferencesOnlyQueriesSelectedProject(t *testing.T) {
	client := &fakeReleaseClient{
		feedbacks: map[string]*api.Feedback{
			"acme/FB-1":  {Reference: "acme/FB-1", Status: "pending_release"},
			"other/FB-2": {Reference: "other/FB-2", Status: "pending_release"},
		},
		failures: map[string]error{},
	}
	committedAt := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

	result := releaseFeedbackReferences(context.Background(), client, []feedbackTrailer{
		{Reference: "acme/FB-1", CommittedAt: committedAt},
		{Reference: "other/FB-2", CommittedAt: committedAt},
	}, "acme")

	if !reflect.DeepEqual(client.requested, []string{"acme/FB-1"}) {
		t.Fatalf("queried feedback %v", client.requested)
	}
	if !reflect.DeepEqual(result.Released, []string{"acme/FB-1"}) {
		t.Fatalf("released %v", result.Released)
	}
	if len(result.Skipped) != 0 || len(result.Failures) != 0 {
		t.Fatalf("unexpected release result: %#v", result)
	}
}

func TestReleaseCommandRequiresProject(t *testing.T) {
	flag := releaseCmd.Flags().Lookup("project")
	if flag == nil {
		t.Fatal("release should expose a --project flag")
	}
	if flag.Annotations[cobra.BashCompOneRequiredFlag][0] != "true" {
		t.Fatal("release should require --project")
	}
}
