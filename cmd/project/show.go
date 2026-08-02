package project

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <handle>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

var showJSON bool

func init() {
	showCmd.Flags().BoolVar(&showJSON, "json", false, "print JSON output")
}

func runShow(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	p, err := client.GetProject(cmd.Context(), args[0])
	if err != nil {
		return fmt.Errorf("getting project: %w", err)
	}

	output.PrintProjectResult(os.Stdout, p, showJSON, "")
	return nil
}
