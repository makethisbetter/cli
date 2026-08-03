package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var declineCmd = &cobra.Command{
	Use:   "decline <handle/FB-n>",
	Short: "Close feedback as not planned",
	Args:  cobra.ExactArgs(1),
	RunE:  runDecline,
}

var declineJSON bool

func init() {
	declineCmd.Flags().BoolVar(&declineJSON, "json", false, "print JSON output")
}

func runDecline(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	feedback, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status:      "closed",
		CloseReason: "not_planned",
	})
	if err != nil {
		return fmt.Errorf("declining feedback: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, feedback, declineJSON,
		fmt.Sprintf("Feedback %s declined.", feedback.Reference))
	return nil
}
