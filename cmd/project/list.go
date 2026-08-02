package project

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	RunE:  runList,
}

var listJSON bool

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "print JSON output")
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects(cmd.Context())
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}

	if listJSON {
		output.PrintJSON(os.Stdout, projects)
	} else {
		output.ProjectTable(os.Stdout, projects)
	}
	return nil
}
