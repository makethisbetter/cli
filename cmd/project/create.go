package project

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

var (
	createDomain string
	createJSON   bool
)

func init() {
	createCmd.Flags().StringVar(&createDomain, "domain", "", "project domain")
	createCmd.Flags().BoolVar(&createJSON, "json", false, "print JSON output")
}

func runCreate(cmd *cobra.Command, args []string) error {
	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	p, err := client.CreateProject(cmd.Context(), api.CreateProjectParams{
		Name:   args[0],
		Domain: createDomain,
	})
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}

	output.PrintProjectResult(os.Stdout, p, createJSON,
		fmt.Sprintf("Project %s created.", p.Name))
	return nil
}
