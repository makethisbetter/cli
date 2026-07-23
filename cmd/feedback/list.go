package feedback

import (
	"fmt"
	"os"
	"sort"

	"github.com/makethisbetter/cli/internal/api"
	"github.com/makethisbetter/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List feedback",
	RunE:  runList,
}

var (
	listStatus   string
	listLabel    string
	listPriority string
	listProject  string
	listSort     string
	listJSON     bool
)

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter by status")
	listCmd.Flags().StringVar(&listLabel, "label", "", "filter by project label")
	listCmd.Flags().StringVar(&listPriority, "priority", "", "filter by priority")
	listCmd.Flags().StringVar(&listProject, "project", "", "project handle")
	listCmd.MarkFlagRequired("project")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort by priority, created, or updated")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "print JSON output")
}

func runList(cmd *cobra.Command, args []string) error {
	if err := validateSort(listSort); err != nil {
		return err
	}

	client, err := loadClientFromConfig()
	if err != nil {
		return err
	}

	feedbacks, err := client.ListFeedbacks(cmd.Context(), api.ListFeedbacksParams{
		Status:        listStatus,
		Label:         listLabel,
		Priority:      listPriority,
		ProjectHandle: listProject,
	})
	if err != nil {
		return fmt.Errorf("listing feedbacks: %w", err)
	}

	sortFeedbacks(feedbacks, listSort)

	if listJSON {
		output.PrintJSON(os.Stdout, feedbacks)
	} else {
		output.FeedbackTable(os.Stdout, feedbacks)
	}
	return nil
}

var validSorts = map[string]bool{
	"":         true,
	"priority": true,
	"created":  true,
	"updated":  true,
}

func validateSort(sortBy string) error {
	if !validSorts[sortBy] {
		return fmt.Errorf("unsupported sort %q (valid: priority, created, updated)", sortBy)
	}
	return nil
}

func sortFeedbacks(feedbacks []api.Feedback, sortBy string) {
	switch sortBy {
	case "priority":
		sort.Slice(feedbacks, func(i, j int) bool {
			return priorityRank(feedbacks[i].Priority) < priorityRank(feedbacks[j].Priority)
		})
	case "created":
		sort.Slice(feedbacks, func(i, j int) bool {
			return feedbacks[i].CreatedAt > feedbacks[j].CreatedAt
		})
	case "updated":
		sort.Slice(feedbacks, func(i, j int) bool {
			return feedbacks[i].UpdatedAt > feedbacks[j].UpdatedAt
		})
	}
}

func priorityRank(priority string) int {
	ranks := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}
	if r, ok := ranks[priority]; ok {
		return r
	}
	return 4
}
