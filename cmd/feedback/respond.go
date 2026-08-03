package feedback

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var respondCmd = &cobra.Command{
	Use:   "respond <handle/FB-n>",
	Short: "Send a response and close received feedback",
	Args:  cobra.ExactArgs(1),
	RunE:  runRespond,
}

var (
	respondBodyFile string
	respondSubject  string
	respondJSON     bool
)

func init() {
	respondCmd.Flags().StringVar(&respondBodyFile, "body-file", "", "UTF-8 response body file, or - for stdin (required)")
	respondCmd.Flags().StringVar(&respondSubject, "subject", "", "email subject (defaults to the Reporter Language)")
	respondCmd.Flags().BoolVar(&respondJSON, "json", false, "print JSON output")
	respondCmd.MarkFlagRequired("body-file")
}

func runRespond(cmd *cobra.Command, args []string) error {
	body, err := readResponseBody(respondBodyFile, cmd.InOrStdin())
	if err != nil {
		return err
	}

	params := api.RespondFeedbackParams{Body: body}
	if cmd.Flags().Changed("subject") {
		subject := strings.TrimSpace(respondSubject)
		if subject == "" {
			return fmt.Errorf("--subject cannot be blank when provided")
		}
		params.Subject = &subject
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}
	result, err := client.RespondFeedback(cmd.Context(), args[0], params)
	if err != nil {
		return fmt.Errorf("responding to feedback: %w", err)
	}

	if respondJSON {
		output.PrintJSON(os.Stdout, result)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Response queued and feedback %s closed.\n", result.Feedback.Reference)
	return nil
}

func readResponseBody(path string, stdin io.Reader) (string, error) {
	var (
		data []byte
		err  error
	)
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("response body must be valid UTF-8")
	}

	body := strings.ReplaceAll(string(data), "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.TrimSpace(body)
	if body == "" {
		return "", fmt.Errorf("response body cannot be blank")
	}
	return body, nil
}
