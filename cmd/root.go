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
