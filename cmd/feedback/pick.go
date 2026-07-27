package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:   "pick <handle/FB-n>",
	Short: "Start working on feedback (set status to in_progress)",
	Args:  cobra.ExactArgs(1),
	RunE:  runPick,
}

var pickJSON bool

func init() {
	pickCmd.Flags().BoolVar(&pickJSON, "json", false, "print JSON output")
}

func runPick(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	fb, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status: "in_progress",
	})
	if err != nil {
		return fmt.Errorf("picking feedback: %w", err)
	}

	output.PrintFeedbackResult(os.Stdout, fb, pickJSON,
		fmt.Sprintf("Feedback %s marked as in_progress.", fb.Reference))
	return nil
}
