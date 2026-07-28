package feedback

import (
	"fmt"
	"os"
	"regexp"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

// bareFeedbackNumberPattern matches the unqualified shorthand users reach for
// when they copy the `id` field out of --json output.
var bareFeedbackNumberPattern = regexp.MustCompile(`^FB-[1-9][0-9]*$`)

var dismissCmd = &cobra.Command{
	Use:   "dismiss <handle/FB-n>",
	Short: "Dismiss feedback (close as not planned)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDismiss,
}

var (
	dismissReason    string
	dismissCanonical string
	dismissJSON      bool
)

var validCloseReasons = map[string]bool{
	"not_planned": true,
	"duplicate":   true,
}

func init() {
	dismissCmd.Flags().StringVar(&dismissReason, "reason", "not_planned", "close reason: not_planned or duplicate")
	dismissCmd.Flags().StringVar(&dismissCanonical, "canonical", "", "canonical feedback reference (required for duplicate)")
	dismissCmd.Flags().BoolVar(&dismissJSON, "json", false, "print JSON output")
}

func runDismiss(cmd *cobra.Command, args []string) error {
	if err := validateDismissOptions(args[0], dismissReason, dismissCanonical); err != nil {
		return err
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	fb, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status:              "closed",
		CloseReason:         dismissReason,
		CanonicalFeedbackID: dismissCanonical,
	})
	if err != nil {
		return fmt.Errorf("dismissing feedback: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, fb, dismissJSON,
		fmt.Sprintf("Feedback %s dismissed (%s).", fb.Reference, dismissReason))
	return nil
}

func validateDismissOptions(reference, reason, canonical string) error {
	if !validCloseReasons[reason] {
		return fmt.Errorf("invalid close reason %q (valid: not_planned, duplicate)", reason)
	}
	if reason == "duplicate" && canonical == "" {
		return fmt.Errorf("--canonical is required when --reason is duplicate")
	}
	if canonical == "" {
		return nil
	}

	targetHandle, _, err := api.ParseFeedbackReference(reference)
	if err != nil {
		return err
	}

	// The server resolves the canonical inside the target's project and drops
	// anything it cannot match, then fails model validation with no useful
	// response. Both a bare FB-n (which is what --json prints as `id`) and a
	// cross-project reference have to be caught here instead.
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
