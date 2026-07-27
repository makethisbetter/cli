package project

import (
	"fmt"
	"os"
	"strings"

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
	createHandle string
	createDomain string
	createJSON   bool
)

func init() {
	createCmd.Flags().StringVar(&createHandle, "handle", "", "globally unique project handle")
	createCmd.MarkFlagRequired("handle")
	createCmd.Flags().StringVar(&createDomain, "domain", "", "domain the widget will run on, e.g. example.com (required)")
	createCmd.MarkFlagRequired("domain")
	createCmd.Flags().BoolVar(&createJSON, "json", false, "print JSON output")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// The server rejects a missing or malformed domain with a bare 422; say
	// what to type instead, before spending a round trip on it.
	domain, err := normalizeDomain(createDomain)
	if err != nil {
		return err
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	p, err := client.CreateProject(cmd.Context(), api.CreateProjectParams{
		Name:   args[0],
		Handle: createHandle,
		Domain: domain,
	})
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}

	output.PrintProjectResult(os.Stdout, p, createJSON,
		fmt.Sprintf("Project %s created.", p.Name))
	return nil
}

// normalizeDomain trims the value and rejects the shapes people actually type
// by mistake: a full URL, a host with a path, or a bare word that is not a
// hostname. Each message names the fix rather than restating the rule.
func normalizeDomain(domain string) (string, error) {
	d := strings.TrimSpace(domain)
	switch {
	case d == "":
		return "", fmt.Errorf("--domain is required: pass the domain the widget will run on, e.g. --domain example.com")
	case strings.Contains(d, "://"):
		host := d[strings.Index(d, "://")+3:]
		host = strings.SplitN(host, "/", 2)[0]
		return "", fmt.Errorf("--domain must be a bare hostname, not a URL: use --domain %s", host)
	case strings.ContainsAny(d, "/ \t"):
		return "", fmt.Errorf("--domain must be a bare hostname with no path or spaces, e.g. example.com (got %q)", d)
	case !strings.Contains(d, "."):
		return "", fmt.Errorf("--domain must be a full hostname including the TLD, e.g. example.com (got %q)", d)
	}
	return d, nil
}
