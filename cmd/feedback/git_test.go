package feedback

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFeedbackReferencesThroughReadsQualifiedTrailers(t *testing.T) {
	dir := t.TempDir()
	runTestGit(t, dir, "init")
	runTestGit(t, dir, "config", "user.name", "Test User")
	runTestGit(t, dir, "config", "user.email", "test@example.com")

	path := filepath.Join(dir, "change.txt")
	if err := os.WriteFile(path, []byte("fixed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, dir, "add", "change.txt")
	runTestGit(t, dir, "commit", "-m", "fix: export", "-m", "Feedback: acme/FB-42\nFeedback: acme/FB-43\nFeedback: invalid")

	references, err := feedbackReferencesThrough(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatalf("reading feedback trailers: %v", err)
	}
	want := []string{"acme/FB-42", "acme/FB-43"}
	if !reflect.DeepEqual(references, want) {
		t.Fatalf("got references %v, want %v", references, want)
	}
}

func TestFeedbackTrailersThroughKeepsLatestCommitTime(t *testing.T) {
	dir := t.TempDir()
	runTestGit(t, dir, "init")
	runTestGit(t, dir, "config", "user.name", "Test User")
	runTestGit(t, dir, "config", "user.email", "test@example.com")

	commitTestFeedback(t, dir, "first", "Feedback: acme/FB-42", "2026-08-01T10:00:00Z")
	commitTestFeedback(t, dir, "second", "Feedback: acme/FB-42", "2026-08-01T11:00:00Z")

	trailers, err := feedbackTrailersThrough(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	wantTime := time.Date(2026, 8, 1, 11, 0, 0, 0, time.UTC)
	if len(trailers) != 1 || trailers[0].Reference != "acme/FB-42" || !trailers[0].CommittedAt.Equal(wantTime) {
		t.Fatalf("unexpected trailers: %#v", trailers)
	}
}

func commitTestFeedback(t *testing.T, dir, contents, trailer, committedAt string) {
	t.Helper()
	path := filepath.Join(dir, "change.txt")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, dir, "add", "change.txt")
	command := exec.Command("git", "-C", dir, "commit", "-m", contents, "-m", trailer)
	command.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+committedAt, "GIT_COMMITTER_DATE="+committedAt)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, output)
	}
}

func TestFeedbackReferencesThroughRejectsUnknownRevision(t *testing.T) {
	dir := t.TempDir()
	runTestGit(t, dir, "init")

	if _, err := feedbackReferencesThrough(context.Background(), dir, "missing"); err == nil {
		t.Fatal("unknown deployed revision should fail")
	}
}

func TestRequireCompleteGitHistoryRejectsShallowClone(t *testing.T) {
	source := t.TempDir()
	runTestGit(t, source, "init")
	runTestGit(t, source, "config", "user.name", "Test User")
	runTestGit(t, source, "config", "user.email", "test@example.com")
	commitTestFeedback(t, source, "first", "Feedback: acme/FB-42", "2026-08-01T10:00:00Z")

	shallow := filepath.Join(t.TempDir(), "shallow")
	command := exec.Command("git", "clone", "--depth", "1", "file://"+source, shallow)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, output)
	}

	err := requireCompleteGitHistory(context.Background(), shallow)
	if err == nil || !strings.Contains(err.Error(), "complete Git history") {
		t.Fatalf("expected shallow history error, got %v", err)
	}
}

func runTestGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
