package output

import (
	"strings"
	"unicode"
)

func columnWidths(rows [][]string, maxWidths []int) []int {
	if len(rows) == 0 {
		return nil
	}
	cols := len(rows[0])
	widths := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if w := displayWidth(cell); i < cols && w > widths[i] {
				widths[i] = w
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
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if displayWidth(s) <= max {
		return s
	}
	if max <= 3 {
		return cutToWidth(s, max)
	}
	return cutToWidth(s, max-3) + "..."
}

// cutToWidth trims s to at most max terminal columns, always on a rune boundary.
// Slicing by byte splits multi-byte runes and the terminal shows U+FFFD instead.
func cutToWidth(s string, max int) string {
	if max <= 0 {
		return ""
	}
	width := 0
	for i, r := range s {
		rw := runeWidth(r)
		if width+rw > max {
			return s[:i]
		}
		width += rw
	}
	return s
}

// displayWidth counts the terminal columns s occupies rather than its bytes, so
// a CJK cell is padded to the same visible width as an ASCII one.
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

func runeWidth(r rune) int {
	switch {
	case unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r):
		return 0
	case isEastAsianWide(r):
		return 2
	default:
		return 1
	}
}

// eastAsianWideRanges are the Unicode TR11 Wide and Fullwidth blocks that show
// up in feedback text: CJK, kana, Hangul, fullwidth forms, and emoji.
var eastAsianWideRanges = [...]struct{ lo, hi rune }{
	{0x1100, 0x115F},   // Hangul Jamo initial consonants
	{0x2E80, 0x303E},   // CJK radicals, Kangxi, CJK symbols and punctuation
	{0x3041, 0x33FF},   // kana, Hangul compatibility Jamo, CJK compatibility
	{0x3400, 0x4DBF},   // CJK unified ideographs extension A
	{0x4E00, 0x9FFF},   // CJK unified ideographs
	{0xA000, 0xA4CF},   // Yi
	{0xAC00, 0xD7A3},   // Hangul syllables
	{0xF900, 0xFAFF},   // CJK compatibility ideographs
	{0xFE30, 0xFE6F},   // CJK compatibility forms, small form variants
	{0xFF00, 0xFF60},   // fullwidth forms
	{0xFFE0, 0xFFE6},   // fullwidth signs
	{0x1F300, 0x1F64F}, // emoji: symbols and pictographs, emoticons
	{0x1F900, 0x1F9FF}, // emoji: supplemental symbols and pictographs
	{0x20000, 0x3FFFD}, // CJK unified ideographs extensions B and beyond
}

func isEastAsianWide(r rune) bool {
	for _, rng := range eastAsianWideRanges {
		if r >= rng.lo && r <= rng.hi {
			return true
		}
	}
	return false
}

func ptrOr(p *string, fallback string) string {
	if p == nil {
		return fallback
	}
	return *p
}
