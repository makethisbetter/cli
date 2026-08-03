package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <handle/FB-n>",
	Short: "Restore one archived feedback",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestore,
}

var restoreJSON bool

func init() {
	restoreCmd.Flags().BoolVar(&restoreJSON, "json", false, "print JSON output")
}

func runRestore(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}
	feedback, err := client.RestoreFeedback(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("restoring feedback: %w", err)
	}

	msg := fmt.Sprintf("Feedback %s is active (not archived).", feedback.Reference)
	if feedback.ArchivedAt != nil {
		msg = fmt.Sprintf("Feedback %s is still archived.", feedback.Reference)
	}
	output.PrintFeedbackResult(os.Stdout, feedback, restoreJSON, msg)
	return nil
}
