package project

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <handle>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

var (
	updateName      string
	updateDomain    string
	updateAIContext string
	updateJSON      bool
)

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "new project name")
	updateCmd.Flags().StringVar(&updateDomain, "domain", "", "domain the widget will run on, e.g. example.com")
	updateCmd.Flags().StringVar(&updateAIContext, "ai-context", "", "project context for AI triage: what the product is, who uses it, key terms")
	updateCmd.Flags().BoolVar(&updateJSON, "json", false, "print JSON output")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Only flags the caller actually set are sent, so an empty --ai-context ""
	// still clears the field while unset flags leave the project untouched.
	var params api.UpdateProjectParams
	if cmd.Flags().Changed("name") {
		params.Name = &updateName
	}
	if cmd.Flags().Changed("domain") {
		domain, err := normalizeDomain(updateDomain)
		if err != nil {
			return err
		}
		params.Domain = &domain
	}
	if cmd.Flags().Changed("ai-context") {
		params.AIContext = &updateAIContext
	}
	if params.Name == nil && params.Domain == nil && params.AIContext == nil {
		return fmt.Errorf("nothing to update: pass at least one of --name, --domain, --ai-context")
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	p, err := client.UpdateProject(cmd.Context(), args[0], params)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}

	output.PrintProjectResult(os.Stdout, p, updateJSON,
		fmt.Sprintf("Project %s updated.", p.Name))
	return nil
}
