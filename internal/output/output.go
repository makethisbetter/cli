package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/makethisbetter/cli/internal/api"
)

var (
	colorOnce    sync.Once
	colorEnabled bool
)

func isColorEnabled() bool {
	colorOnce.Do(func() {
		if os.Getenv("NO_COLOR") != "" {
			colorEnabled = false
			return
		}
		fi, err := os.Stdout.Stat()
		if err != nil {
			colorEnabled = false
			return
		}
		colorEnabled = fi.Mode()&os.ModeCharDevice != 0
	})
	return colorEnabled
}

func wrap(code, text string) string {
	if !isColorEnabled() {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", code, text)
}

// Green wraps text in the green ANSI color when color output is enabled.
func Green(text string) string { return wrap("32", text) }

// Red wraps text in the red ANSI color when color output is enabled.
func Red(text string) string { return wrap("31", text) }

// Bold wraps text in the bold ANSI style when color output is enabled.
func Bold(text string) string { return wrap("1", text) }

// Dim wraps text in the dim ANSI style when color output is enabled.
func Dim(text string) string { return wrap("2", text) }

// Yellow wraps text in the yellow ANSI color when color output is enabled.
func Yellow(text string) string { return wrap("33", text) }

// ColorStatus returns status colored according to its value.
func ColorStatus(status string) string {
	switch status {
	case "in_progress":
		return Green(status)
	case "closed":
		return Dim(status)
	default:
		return status
	}
}

// PrintJSON writes v to w as indented JSON.
func PrintJSON(w io.Writer, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return
	}
	fmt.Fprintln(w, string(data))
}

// FeedbackTable writes feedbacks to w as an aligned table.
func FeedbackTable(w io.Writer, feedbacks []api.Feedback) {
	if len(feedbacks) == 0 {
		fmt.Fprintln(w, "No feedback found.")
		return
	}

	rows := make([][]string, 0, len(feedbacks)+1)
	rows = append(rows, []string{"ID", "Status", "Priority", "Type", "Description", "Updated"})
	for _, fb := range feedbacks {
		rows = append(rows, []string{
			fb.Reference,
			fb.Status,
			fb.Priority,
			ptrOr(fb.FeedbackType, ""),
			truncate(fb.Description, 48),
			formatDate(fb.UpdatedAt),
		})
	}

	widths := columnWidths(rows, []int{0, 0, 0, 0, 48, 0})
	for i, row := range rows {
		line := formatRow(row, widths)
		if i == 0 {
			fmt.Fprintln(w, Bold(line))
			fmt.Fprintln(w, Dim(separatorLine(widths)))
		} else {
			fmt.Fprintln(w, line)
		}
	}
}

// FeedbackDetail writes a full single-feedback view to w.
func FeedbackDetail(w io.Writer, fb *api.Feedback) {
	fmt.Fprintln(w, Bold(fmt.Sprintf("%s  %s  %s", fb.Reference, ColorStatus(fb.Status), fb.Priority)))
	fmt.Fprintln(w)
	fmt.Fprintln(w, fb.Description)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", Dim("Project:"), fb.ProjectID)
	fmt.Fprintf(w, "%s %s\n", Dim("Type:"), ptrOr(fb.FeedbackType, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Recommendation:"), ptrOr(fb.Recommendation, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Close reason:"), ptrOr(fb.CloseReason, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Page:"), ptrOr(fb.PageURL, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Reporter:"), formatReporter(fb))
	fmt.Fprintf(w, "%s %s\n", Dim("Created:"), formatDate(fb.CreatedAt))
	fmt.Fprintf(w, "%s %s\n", Dim("Updated:"), formatDate(fb.UpdatedAt))

	if fb.ScreenshotAttached {
		fmt.Fprintf(w, "%s %s\n", Dim("Screenshot:"), Green("attached"))
	}
	if fb.RecordingAttached {
		dur := "-"
		if fb.RecordingDuration != nil {
			dur = fmt.Sprintf("%ds", *fb.RecordingDuration)
		}
		fmt.Fprintf(w, "%s %s (%s)\n", Dim("Recording:"), Green("attached"), dur)
	}

	if len(fb.AISummary) > 0 && string(fb.AISummary) != "null" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, Bold("AI analysis"))
		printIndentedJSON(w, fb.AISummary)
	}

	if len(fb.TargetElement) > 0 && string(fb.TargetElement) != "null" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, Bold("Target element"))
		printIndentedJSON(w, fb.TargetElement)
	}

	if len(fb.ConsoleErrors) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, Bold("Console errors"))
		for _, e := range fb.ConsoleErrors {
			fmt.Fprintf(w, "- %s\n", string(e))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, Bold("Timeline"))
	fmt.Fprintf(w, "- received: %s\n", formatDate(fb.CreatedAt))
	if fb.Status != "received" {
		fmt.Fprintf(w, "- %s: %s\n", fb.Status, formatDate(fb.UpdatedAt))
	}
}

// PrintFeedbackResult handles the shared JSON-vs-formatted output pattern
// used by pick, resolve, dismiss, and show commands.
func PrintFeedbackResult(w io.Writer, fb *api.Feedback, jsonOutput bool, msg string) {
	if jsonOutput {
		PrintJSON(w, fb)
		return
	}
	if msg != "" {
		fmt.Fprintf(w, "%s %s\n\n", Green("*"), msg)
	}
	FeedbackDetail(w, fb)
}

// ProjectTable writes projects to w as an aligned table.
func ProjectTable(w io.Writer, projects []api.Project) {
	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects found.")
		return
	}

	rows := make([][]string, 0, len(projects)+1)
	rows = append(rows, []string{"ID", "Name", "Feedbacks", "Created"})
	for _, p := range projects {
		rows = append(rows, []string{
			p.ID,
			p.Name,
			strconv.Itoa(p.FeedbacksCount),
			formatDate(p.CreatedAt),
		})
	}

	widths := columnWidths(rows, []int{0, 0, 0, 0})
	for i, row := range rows {
		line := formatRow(row, widths)
		if i == 0 {
			fmt.Fprintln(w, Bold(line))
			fmt.Fprintln(w, Dim(separatorLine(widths)))
		} else {
			fmt.Fprintln(w, line)
		}
	}
}

// ProjectDetail writes a full single-project view to w.
func ProjectDetail(w io.Writer, p *api.Project) {
	fmt.Fprintln(w, Bold(fmt.Sprintf("%s  %s", p.Name, p.ID)))
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", Dim("API key:"), p.APIKey)
	fmt.Fprintf(w, "%s %s\n", Dim("Board URL:"), ptrOr(p.BoardURL, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Domain:"), ptrOr(p.Domain, "-"))
	fmt.Fprintf(w, "%s %s\n", Dim("Feedback visibility:"), p.FeedbackVisibility)
	fmt.Fprintf(w, "%s %t\n", Dim("Identity verification:"), p.EnforceIdentityVerification)
	if p.SigningSecret != nil {
		fmt.Fprintf(w, "%s %s\n", Dim("Signing secret:"), *p.SigningSecret)
	} else {
		fmt.Fprintf(w, "%s %s\n", Dim("Signing secret:"), Dim("(admin only)"))
	}
	fmt.Fprintf(w, "%s %d\n", Dim("Feedbacks:"), p.FeedbacksCount)
	fmt.Fprintf(w, "%s %s\n", Dim("Created:"), formatDate(p.CreatedAt))
	fmt.Fprintf(w, "%s %s\n", Dim("Updated:"), formatDate(p.UpdatedAt))
}

// PrintProjectResult handles the shared JSON-vs-formatted output pattern
// used by the project list, show, and create commands.
func PrintProjectResult(w io.Writer, p *api.Project, jsonOutput bool, msg string) {
	if jsonOutput {
		PrintJSON(w, p)
		return
	}
	if msg != "" {
		fmt.Fprintf(w, "%s %s\n\n", Green("*"), msg)
	}
	ProjectDetail(w, p)
}

// Error writes a red error message to stderr.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, Red("Error: ")+msg)
}

// Success writes a green success message to w.
func Success(w io.Writer, msg string) {
	fmt.Fprintln(w, Green(msg))
}

func formatReporter(fb *api.Feedback) string {
	var parts []string
	if fb.ReporterName != nil {
		parts = append(parts, *fb.ReporterName)
	}
	if fb.ReporterEmail != nil {
		parts = append(parts, *fb.ReporterEmail)
	}
	if len(parts) == 0 {
		return "-"
	}
	if len(parts) == 2 {
		return parts[0] + " <" + parts[1] + ">"
	}
	return parts[0]
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return s
		}
	}
	return t.Local().Format("2006/01/02 15:04")
}

func printIndentedJSON(w io.Writer, raw json.RawMessage) {
	var v any
	if json.Unmarshal(raw, &v) == nil {
		data, err := json.MarshalIndent(v, "", "  ")
		if err == nil {
			fmt.Fprintln(w, string(data))
			return
		}
	}
	fmt.Fprintln(w, string(raw))
}
