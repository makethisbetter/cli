package cmd

import (
	"github.com/makethisbetter/cli/cmd/feedback"
	"github.com/makethisbetter/cli/cmd/project"
	"github.com/spf13/cobra"
)

// version is injected at build time via -ldflags "-X github.com/makethisbetter/cli/cmd.version=x.y.z".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "makethisbetter",
	Short:   "Manage Make This Better feedback from the terminal",
	Version: version,
	// Cobra prints the usage block after any error, which buries a network or API
	// failure under two screens of flag help. Usage only helps with invocation
	// mistakes, so silence it once the invocation itself has been accepted.
	// Required-flag validation normally runs after this hook, so run it here to
	// keep "required flag not set" on the usage-printing side of the line.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if err := cmd.ValidateRequiredFlags(); err != nil {
			return err
		}
		if err := cmd.ValidateFlagGroups(); err != nil {
			return err
		}
		cmd.Root().SilenceUsage = true
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(feedback.Cmd)
	rootCmd.AddCommand(project.Cmd)
}

func Execute() error {
	return rootCmd.Execute()
}
