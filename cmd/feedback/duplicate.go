package feedback

import (
	"fmt"
	"os"
	"regexp"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var bareFeedbackNumberPattern = regexp.MustCompile(`^FB-[1-9][0-9]*$`)

var duplicateCmd = &cobra.Command{
	Use:   "duplicate <handle/FB-n>",
	Short: "Close feedback as a duplicate",
	Args:  cobra.ExactArgs(1),
	RunE:  runDuplicate,
}

var (
	duplicateCanonical string
	duplicateJSON      bool
)

func init() {
	duplicateCmd.Flags().StringVar(&duplicateCanonical, "canonical", "", "canonical feedback reference (required)")
	duplicateCmd.Flags().BoolVar(&duplicateJSON, "json", false, "print JSON output")
	duplicateCmd.MarkFlagRequired("canonical")
}

func runDuplicate(cmd *cobra.Command, args []string) error {
	if err := validateCanonicalReference(args[0], duplicateCanonical); err != nil {
		return err
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	feedback, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status:              "closed",
		CloseReason:         "duplicate",
		CanonicalFeedbackID: duplicateCanonical,
	})
	if err != nil {
		return fmt.Errorf("marking feedback duplicate: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, feedback, duplicateJSON,
		fmt.Sprintf("Feedback %s marked as a duplicate of %s.", feedback.Reference, duplicateCanonical))
	return nil
}

func validateCanonicalReference(reference, canonical string) error {
	if canonical == "" {
		return fmt.Errorf("--canonical is required")
	}

	targetHandle, _, err := api.ParseFeedbackReference(reference)
	if err != nil {
		return err
	}
	if bareFeedbackNumberPattern.MatchString(canonical) {
		return fmt.Errorf("--canonical must include the project handle, use %s/%s", targetHandle, canonical)
	}
	canonicalHandle, _, err := api.ParseFeedbackReference(canonical)
	if err != nil {
		return fmt.Errorf("invalid --canonical %q: %w", canonical, err)
	}
	if canonicalHandle != targetHandle {
		return fmt.Errorf("--canonical %s is in another project, a duplicate's canonical must be in %s", canonical, targetHandle)
	}
	return nil
}
