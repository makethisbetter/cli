package cmd

import (
	"fmt"
	"os"

	"github.com/makethisbetter/cli/internal/config"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current connection status",
	RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	w := os.Stdout
	fmt.Fprintln(w, output.Bold("Make This Better CLI"))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", output.Dim("API URL:"), cfg.APIURL)

	if cfg.Token != "" {
		maskedToken := cfg.Token
		if len(maskedToken) > 8 {
			maskedToken = maskedToken[:4] + "..." + maskedToken[len(maskedToken)-4:]
		}
		fmt.Fprintf(w, "%s %s\n", output.Dim("Token:"), maskedToken)
	} else {
		fmt.Fprintf(w, "%s %s\n", output.Dim("Token:"), output.Red("not set"))
	}

	if cfg.UserEmail != "" {
		fmt.Fprintf(w, "%s %s\n", output.Dim("User:"), cfg.UserEmail)
	}

	if cfg.AccountID != "" {
		fmt.Fprintf(w, "%s %s\n", output.Dim("Account:"), cfg.AccountID)
	}

	if cfg.Token == "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, output.Yellow("Run `makethisbetter login` to authenticate."))
	}

	return nil
}
