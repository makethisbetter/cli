package feedback

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <handle/FB-n>",
	Short: "Resolve feedback (close as shipped)",
	Args:  cobra.ExactArgs(1),
	RunE:  runResolve,
}

var (
	resolvePR   string
	resolveJSON bool
)

func init() {
	resolveCmd.Flags().StringVar(&resolvePR, "pr", "", "pull request URL")
	resolveCmd.Flags().BoolVar(&resolveJSON, "json", false, "print JSON output")
}

func runResolve(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	labels := map[string]any{
		"close_reason": "shipped",
	}
	if resolvePR != "" {
		labels["pr_url"] = resolvePR
	}

	fb, err := client.UpdateFeedback(cmd.Context(), args[0], api.UpdateFeedbackParams{
		Status: "closed",
		Labels: labels,
	})
	if err != nil {
		return fmt.Errorf("resolving feedback: %w", err)
	}

	msg := fmt.Sprintf("Feedback %s resolved (shipped).", fb.Reference)
	if resolvePR != "" {
		msg += fmt.Sprintf(" PR: %s", resolvePR)
	}
	output.PrintFeedbackResult(os.Stdout, fb, resolveJSON, msg)
	return nil
}
