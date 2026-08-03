package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <handle/FB-n>",
	Short: "Archive one unclaimed feedback",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchive,
}

var archiveJSON bool

func init() {
	archiveCmd.Flags().BoolVar(&archiveJSON, "json", false, "print JSON output")
}

func runArchive(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}
	feedback, err := client.ArchiveFeedback(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("archiving feedback: %w", err)
	}

	msg := fmt.Sprintf("Feedback %s is archived.", feedback.Reference)
	if feedback.ArchivedAt == nil {
		msg = fmt.Sprintf("Feedback %s is not archived.", feedback.Reference)
	}
	output.PrintFeedbackResult(os.Stdout, feedback, archiveJSON, msg)
	return nil
}
