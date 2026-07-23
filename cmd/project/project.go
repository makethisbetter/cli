package project

import (
	"fmt"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/config"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(createCmd)
}

func loadClientFromConfig() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if _, err := config.RequireToken(cfg); err != nil {
		return nil, err
	}
	return api.NewClient(cfg), nil
}
