package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen <handle/FB-n>",
	Short: "Reopen closed feedback for a new work cycle",
	Args:  cobra.ExactArgs(1),
	RunE:  runReopen,
}

var reopenJSON bool

func init() {
	reopenCmd.Flags().BoolVar(&reopenJSON, "json", false, "print JSON output")
}

func runReopen(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	feedback, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{Status: "received"})
	if err != nil {
		return fmt.Errorf("reopening feedback: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, feedback, reopenJSON,
		fmt.Sprintf("Feedback %s reopened.", feedback.Reference))
	return nil
}
