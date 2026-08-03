package feedback

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/makethisbetter/cli/internal/api"
)

type feedbackTrailer struct {
	Reference   string    `json:"reference"`
	CommittedAt time.Time `json:"committed_at"`
}

func feedbackReferencesThrough(ctx context.Context, dir, revision string) ([]string, error) {
	trailers, err := feedbackTrailersThrough(ctx, dir, revision)
	if err != nil {
		return nil, err
	}

	references := make([]string, 0, len(trailers))
	for _, trailer := range trailers {
		references = append(references, trailer.Reference)
	}
	return references, nil
}

func feedbackTrailersThrough(ctx context.Context, dir, revision string) ([]feedbackTrailer, error) {
	resolved, err := gitOutput(ctx, dir, "rev-parse", "--verify", "--end-of-options", revision+"^{commit}")
	if err != nil {
		return nil, fmt.Errorf("deployed revision %q is not a commit: %w", revision, err)
	}

	output, err := gitOutput(
		ctx,
		dir,
		"log",
		strings.TrimSpace(resolved),
		"--format=%cI%x09%(trailers:key=Feedback,valueonly,separator=%x2c)",
	)
	if err != nil {
		return nil, fmt.Errorf("reading Feedback trailers through %q: %w", revision, err)
	}

	latest := make(map[string]feedbackTrailer)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "\t", 2)
		if len(parts) != 2 {
			continue
		}
		committedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("reading Feedback trailer commit time: %w", err)
		}
		for _, rawReference := range strings.Split(parts[1], ",") {
			reference := strings.TrimSpace(rawReference)
			if _, _, err := api.ParseFeedbackReference(reference); err == nil {
				if _, exists := latest[reference]; !exists {
					latest[reference] = feedbackTrailer{Reference: reference, CommittedAt: committedAt}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading Feedback trailers: %w", err)
	}

	trailers := make([]feedbackTrailer, 0, len(latest))
	for _, trailer := range latest {
		trailers = append(trailers, trailer)
	}
	sort.Slice(trailers, func(i, j int) bool { return trailers[i].Reference < trailers[j].Reference })
	return trailers, nil
}

func requireCompleteGitHistory(ctx context.Context, dir string) error {
	shallow, err := gitOutput(ctx, dir, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return fmt.Errorf("checking Git history completeness: %w", err)
	}
	if strings.TrimSpace(shallow) == "true" {
		return fmt.Errorf("release requires complete Git history; fetch the full history before retrying")
	}
	return nil
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", dir}, args...)
	output, err := exec.CommandContext(ctx, "git", commandArgs...).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return string(output), nil
}
