package feedback

import (
	"fmt"
	"os"
	"strings"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var readyCmd = &cobra.Command{
	Use:   "ready <handle/FB-n>",
	Short: "Mark implemented feedback ready for release",
	Args:  cobra.ExactArgs(1),
	RunE:  runReady,
}

var (
	readySummary string
	readyJSON    bool
)

func init() {
	readyCmd.Flags().StringVar(&readySummary, "summary", "", "factual summary of what changed (required)")
	readyCmd.Flags().BoolVar(&readyJSON, "json", false, "print JSON output")
}

func runReady(cmd *cobra.Command, args []string) error {
	if err := validateReadyOptions(readySummary); err != nil {
		return err
	}

	reference := args[0]
	references, err := feedbackReferencesThrough(cmd.Context(), ".", "HEAD")
	if err != nil {
		return err
	}
	if !hasFeedbackReference(references, reference) {
		return fmt.Errorf("current Git history does not contain a `Feedback: %s` commit trailer", reference)
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	feedback, err := client.ReadyFeedback(cmd.Context(), reference, strings.TrimSpace(readySummary))
	if err != nil {
		return fmt.Errorf("marking feedback ready: %w", err)
	}

	message := fmt.Sprintf("Feedback %s is ready for release.", feedback.Reference)
	output.PrintFeedbackResult(os.Stdout, feedback, readyJSON, message)
	return nil
}

func validateReadyOptions(summary string) error {
	if strings.TrimSpace(summary) == "" {
		return fmt.Errorf("--summary is required")
	}
	return nil
}

func hasFeedbackReference(references []string, wanted string) bool {
	for _, reference := range references {
		if reference == wanted {
			return true
		}
	}
	return false
}
