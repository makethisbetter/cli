package feedback

import (
	"testing"

	"github.com/makethisbetter/cli/internal/api"
)

func TestListCommandKeepsLegacyTypeFlag(t *testing.T) {
	if listCmd.Flags().Lookup("type") == nil {
		t.Fatal("expected --type to remain available as a compatibility alias")
	}
}

func TestSelectedListLabel(t *testing.T) {
	tests := []struct {
		name       string
		label      string
		legacyType string
		want       string
		wantErr    bool
	}{
		{name: "label", label: "Safari", want: "Safari"},
		{name: "legacy type", legacyType: "bug", want: "bug"},
		{name: "no filter"},
		{name: "conflicting flags", label: "Safari", legacyType: "bug", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectedListLabel(tt.label, tt.legacyType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("selectedListLabel(%q, %q) error = %v, wantErr %v", tt.label, tt.legacyType, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("selectedListLabel(%q, %q) = %q, want %q", tt.label, tt.legacyType, got, tt.want)
			}
		})
	}
}

func TestValidateSort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty string is valid", input: "", wantErr: false},
		{name: "priority is valid", input: "priority", wantErr: false},
		{name: "created is valid", input: "created", wantErr: false},
		{name: "updated is valid", input: "updated", wantErr: false},
		{name: "unknown value is invalid", input: "name", wantErr: true},
		{name: "uppercase is invalid", input: "Priority", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSort(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSortFeedbacksByPriority(t *testing.T) {
	feedbacks := []api.Feedback{
		{ID: "1", Priority: "low"},
		{ID: "2", Priority: "critical"},
		{ID: "3", Priority: "high"},
		{ID: "4", Priority: "medium"},
	}

	sortFeedbacks(feedbacks, "priority")

	want := []string{"critical", "high", "medium", "low"}
	for i, w := range want {
		if feedbacks[i].Priority != w {
			t.Errorf("index %d: got priority %q, want %q", i, feedbacks[i].Priority, w)
		}
	}
}

func TestSortFeedbacksByCreated(t *testing.T) {
	feedbacks := []api.Feedback{
		{ID: "old", CreatedAt: "2024-01-01T00:00:00Z"},
		{ID: "new", CreatedAt: "2024-06-01T00:00:00Z"},
		{ID: "mid", CreatedAt: "2024-03-01T00:00:00Z"},
	}

	sortFeedbacks(feedbacks, "created")

	want := []string{"new", "mid", "old"}
	for i, w := range want {
		if feedbacks[i].ID != w {
			t.Errorf("index %d: got ID %q, want %q", i, feedbacks[i].ID, w)
		}
	}
}

func TestSortFeedbacksByUpdated(t *testing.T) {
	feedbacks := []api.Feedback{
		{ID: "stale", UpdatedAt: "2024-01-01T00:00:00Z"},
		{ID: "fresh", UpdatedAt: "2024-06-01T00:00:00Z"},
	}

	sortFeedbacks(feedbacks, "updated")

	if feedbacks[0].ID != "fresh" {
		t.Errorf("expected fresh first, got %q", feedbacks[0].ID)
	}
	if feedbacks[1].ID != "stale" {
		t.Errorf("expected stale second, got %q", feedbacks[1].ID)
	}
}

func TestPriorityRank(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		want     int
	}{
		{name: "critical is 0", priority: "critical", want: 0},
		{name: "high is 1", priority: "high", want: 1},
		{name: "medium is 2", priority: "medium", want: 2},
		{name: "low is 3", priority: "low", want: 3},
		{name: "unknown falls to 4", priority: "none", want: 4},
		{name: "empty falls to 4", priority: "", want: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityRank(tt.priority)
			if got != tt.want {
				t.Errorf("priorityRank(%q) = %d, want %d", tt.priority, got, tt.want)
			}
		})
	}
}

func TestPriorityRankOrdering(t *testing.T) {
	if priorityRank("critical") >= priorityRank("high") {
		t.Error("critical should rank before high")
	}
	if priorityRank("high") >= priorityRank("medium") {
		t.Error("high should rank before medium")
	}
	if priorityRank("medium") >= priorityRank("low") {
		t.Error("medium should rank before low")
	}
	if priorityRank("low") >= priorityRank("unknown") {
		t.Error("low should rank before unknown")
	}
}
