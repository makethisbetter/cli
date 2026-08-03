package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

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
	loginSendOnly  bool
)

func init() {
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "email address for OTP login")
	loginCmd.Flags().StringVar(&loginOTP, "otp", "", "six digit verification code")
	loginCmd.Flags().StringVar(&loginAPIURL, "api-url", config.DefaultAPIURL, "Make This Better API URL")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "save an existing API token without OTP")
	loginCmd.Flags().StringVar(&loginAccountID, "account-id", "", "default account id for API requests")
	loginCmd.Flags().BoolVar(&loginSendOnly, "send-only", false, "send a verification code without waiting for input")
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

	if loginOTP != "" {
		if loginEmail != "" || loginSendOnly {
			return fmt.Errorf("--otp cannot be combined with --email or --send-only")
		}
		return completePendingLogin(cmd, apiURL)
	}
	if loginSendOnly && strings.TrimSpace(loginEmail) == "" {
		return fmt.Errorf("--email is required with --send-only")
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
	if loginSendOnly {
		pending := &config.PendingLogin{
			Email:             email,
			APIURL:            apiURL,
			RegistrationToken: reg.RegistrationToken,
			ExpiresAt:         time.Now().Add(time.Duration(reg.ExpiresIn) * time.Second),
		}
		if err := config.SavePendingLogin(pending); err != nil {
			return err
		}
		fmt.Println("Run `makethisbetter login --otp <code>` to complete login.")
		return nil
	}

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

func completePendingLogin(cmd *cobra.Command, requestedAPIURL string) error {
	pending, err := config.LoadPendingLogin()
	if err != nil {
		return err
	}
	if !pending.ExpiresAt.After(time.Now()) {
		if err := config.RemovePendingLogin(); err != nil {
			return err
		}
		return fmt.Errorf("pending login has expired, request a new verification code")
	}
	if cmd.Flags().Changed("api-url") && requestedAPIURL != pending.APIURL {
		return fmt.Errorf("--api-url does not match the pending login")
	}

	client := api.NewUnauthClient(pending.APIURL)
	result, err := client.VerifyRegistration(cmd.Context(), pending.RegistrationToken, strings.TrimSpace(loginOTP))
	if err != nil {
		return fmt.Errorf("verifying code: %w", err)
	}

	accountID := loginAccountID
	if accountID == "" {
		accountID = result.Account.ID
	}
	cfg := &config.Config{
		Token:     result.APIToken.Token,
		APIURL:    pending.APIURL,
		AccountID: accountID,
		UserEmail: result.User.Email,
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	if err := config.RemovePendingLogin(); err != nil {
		return err
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
