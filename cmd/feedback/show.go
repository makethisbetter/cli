package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show feedback details",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

var (
	showJSON bool
	showMD   bool
)

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "print JSON output")
	showCmd.Flags().BoolVar(&showMD, "md", false, "print server-rendered markdown (agent-friendly)")
}

func runShow(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	fb, err := client.GetFeedback(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("getting feedback: %w", err)
	}

	if showMD && !showJSON {
		if fb.Markdown == "" {
			return fmt.Errorf("server did not return markdown (self-hosted instance may predate this feature); use --json or upgrade the server")
		}
		fmt.Fprintln(os.Stdout, fb.Markdown)
		return nil
	}

	output.PrintFeedbackResult(os.Stdout, fb, showJSON, "")
	return nil
}
