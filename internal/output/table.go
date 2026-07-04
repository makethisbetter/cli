package output

import "strings"

func columnWidths(rows [][]string, maxWidths []int) []int {
	if len(rows) == 0 {
		return nil
	}
	cols := len(rows[0])
	widths := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if i < cols && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	for i, mw := range maxWidths {
		if i < cols && mw > 0 && widths[i] > mw {
			widths[i] = mw
		}
	}
	return widths
}

func formatRow(row []string, widths []int) string {
	parts := make([]string, len(row))
	for i, cell := range row {
		w := 0
		if i < len(widths) {
			w = widths[i]
		}
		parts[i] = padRight(cell, w)
	}
	return strings.TrimRight(strings.Join(parts, "  "), " ")
}

func separatorLine(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w)
	}
	return strings.Join(parts, "  ")
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func ptrOr(p *string, fallback string) string {
	if p == nil {
		return fallback
	}
	return *p
}
