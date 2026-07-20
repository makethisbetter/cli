package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var dismissCmd = &cobra.Command{
	Use:   "dismiss <handle/FB-n>",
	Short: "Dismiss feedback (close as not planned)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDismiss,
}

var (
	dismissReason string
	dismissJSON   bool
)

var validCloseReasons = map[string]bool{
	"not_planned": true,
	"duplicate":   true,
}

func init() {
	dismissCmd.Flags().StringVar(&dismissReason, "reason", "not_planned", "close reason: not_planned or duplicate")
	dismissCmd.Flags().BoolVar(&dismissJSON, "json", false, "print JSON output")
}

func runDismiss(cmd *cobra.Command, args []string) error {
	if !validCloseReasons[dismissReason] {
		return fmt.Errorf("invalid close reason %q (valid: not_planned, duplicate)", dismissReason)
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	fb, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status: "closed",
		Labels: map[string]any{
			"close_reason": dismissReason,
		},
	})
	if err != nil {
		return fmt.Errorf("dismissing feedback: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, fb, dismissJSON,
		fmt.Sprintf("Feedback %s dismissed (%s).", fb.Reference, dismissReason))
	return nil
}
