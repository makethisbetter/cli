package feedback

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release feedback included in a deployed Git revision",
	Args:  cobra.NoArgs,
	RunE:  runRelease,
}

var (
	releaseThrough string
	releaseJSON    bool
)

func init() {
	releaseCmd.Flags().StringVar(&releaseThrough, "through", "", "deployed Git commit SHA (required)")
	releaseCmd.Flags().BoolVar(&releaseJSON, "json", false, "print JSON output")
	releaseCmd.MarkFlagRequired("through")
}

type releaseClient interface {
	GetFeedback(context.Context, string) (*api.Feedback, error)
	ReleaseFeedback(context.Context, string, string) (*api.Feedback, error)
}

type releaseFailure struct {
	Reference string `json:"reference"`
	Error     string `json:"error"`
}

type releaseSkip struct {
	Reference string `json:"reference"`
	Reason    string `json:"reason"`
}

type releaseResult struct {
	Released []string         `json:"released"`
	Skipped  []releaseSkip    `json:"skipped"`
	Failures []releaseFailure `json:"failures"`
}

func runRelease(cmd *cobra.Command, _ []string) error {
	if err := requireCompleteGitHistory(cmd.Context(), "."); err != nil {
		return err
	}
	trailers, err := feedbackTrailersThrough(cmd.Context(), ".", releaseThrough)
	if err != nil {
		return err
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}
	result := releaseFeedbackReferences(cmd.Context(), client, trailers)

	if releaseJSON {
		output.PrintJSON(os.Stdout, result)
	} else {
		printReleaseResult(result)
	}
	if len(result.Failures) > 0 {
		return fmt.Errorf("failed to release %d feedback item(s)", len(result.Failures))
	}
	return nil
}

func releaseFeedbackReferences(ctx context.Context, client releaseClient, trailers []feedbackTrailer) releaseResult {
	result := releaseResult{
		Released: []string{},
		Skipped:  []releaseSkip{},
		Failures: []releaseFailure{},
	}

	for _, trailer := range trailers {
		reference := trailer.Reference
		feedback, err := client.GetFeedback(ctx, reference)
		if err != nil {
			var apiError *api.APIError
			if errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound {
				result.Skipped = append(result.Skipped, releaseSkip{Reference: reference, Reason: "not_found"})
				continue
			}
			result.Failures = append(result.Failures, releaseFailure{Reference: reference, Error: err.Error()})
			continue
		}
		if feedback.Status != "pending_release" {
			result.Skipped = append(result.Skipped, releaseSkip{Reference: reference, Reason: "not_pending_release"})
			continue
		}
		if _, err := client.ReleaseFeedback(ctx, reference, trailer.CommittedAt.Format(time.RFC3339)); err != nil {
			var apiError *api.APIError
			if errors.As(err, &apiError) && apiError.StatusCode == http.StatusConflict {
				result.Skipped = append(result.Skipped, releaseSkip{Reference: reference, Reason: "stale_trailer_before_reopen"})
				continue
			}
			result.Failures = append(result.Failures, releaseFailure{Reference: reference, Error: err.Error()})
			continue
		}
		result.Released = append(result.Released, reference)
	}

	return result
}

func printReleaseResult(result releaseResult) {
	fmt.Fprintf(os.Stdout, "Released: %s\n", releaseList(result.Released))
	fmt.Fprintf(os.Stdout, "Skipped: %s\n", releaseSkipList(result.Skipped))
	for _, failure := range result.Failures {
		fmt.Fprintf(os.Stdout, "Failed: %s (%s)\n", failure.Reference, failure.Error)
	}
}

func releaseSkipList(skips []releaseSkip) string {
	if len(skips) == 0 {
		return "none"
	}
	values := make([]string, 0, len(skips))
	for _, skip := range skips {
		values = append(values, fmt.Sprintf("%s (%s)", skip.Reference, skip.Reason))
	}
	return strings.Join(values, ", ")
}

func releaseList(references []string) string {
	if len(references) == 0 {
		return "none"
	}
	return strings.Join(references, ", ")
}
