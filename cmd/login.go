package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/config"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with an email verification code",
	RunE:  runLogin,
}

var (
	loginEmail     string
	loginOTP       string
	loginAPIURL    string
	loginToken     string
	loginAccountID string
)

func init() {
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "email address for OTP login")
	loginCmd.Flags().StringVar(&loginOTP, "otp", "", "six digit verification code")
	loginCmd.Flags().StringVar(&loginAPIURL, "api-url", config.DefaultAPIURL, "Make This Better API URL")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "save an existing API token without OTP")
	loginCmd.Flags().StringVar(&loginAccountID, "account-id", "", "default account id for API requests")
}

func runLogin(cmd *cobra.Command, args []string) error {
	apiURL := config.NormalizeURL(loginAPIURL)

	if loginToken != "" {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load existing config: %v\n", err)
			cfg = &config.Config{APIURL: apiURL}
		}
		cfg.Token = loginToken
		cfg.APIURL = apiURL
		if loginAccountID != "" {
			cfg.AccountID = loginAccountID
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		output.Success(os.Stdout, "Saved API token.")
		return nil
	}

	email := loginEmail
	if email == "" {
		var err error
		email, err = prompt("Email: ")
		if err != nil {
			return err
		}
	}
	email = strings.TrimSpace(email)

	client := api.NewUnauthClient(apiURL)
	reg, err := client.RequestRegistration(cmd.Context(), email)
	if err != nil {
		return fmt.Errorf("requesting verification code: %w", err)
	}

	fmt.Printf("Verification code sent to %s.\n", email)

	otp := loginOTP
	if otp == "" {
		otp, err = prompt("Verification code: ")
		if err != nil {
			return err
		}
	}
	otp = strings.TrimSpace(otp)

	result, err := client.VerifyRegistration(cmd.Context(), reg.RegistrationToken, otp)
	if err != nil {
		return fmt.Errorf("verifying code: %w", err)
	}

	accountID := loginAccountID
	if accountID == "" {
		accountID = result.Account.ID
	}

	cfg := &config.Config{
		Token:     result.APIToken.Token,
		APIURL:    apiURL,
		AccountID: accountID,
		UserEmail: result.User.Email,
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	output.Success(os.Stdout, fmt.Sprintf("Logged in as %s.", result.User.Email))
	return nil
}

func prompt(label string) (string, error) {
	fmt.Print(label)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no input received")
	}
	return scanner.Text(), nil
}
