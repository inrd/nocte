package app

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func editorSizeStatus(content string) string {
	return fmt.Sprintf("Size %s", humanSize(int64(len([]byte(content)))))
}

func largeNoteWarning(sizeBytes int64) string {
	if sizeBytes <= largeNoteWarningThreshold {
		return ""
	}

	return fmt.Sprintf(
		"Large note warning: %s may slow editing and saving.",
		humanSize(sizeBytes),
	)
}

func formatNoteUpdatedAt(modTime time.Time) string {
	if modTime.IsZero() {
		return "Unknown"
	}

	now := time.Now().In(modTime.Location())
	if sameDay(modTime, now) {
		return fmt.Sprintf("Today %s", modTime.Format("15:04"))
	}

	if sameDay(modTime, now.AddDate(0, 0, -1)) {
		return fmt.Sprintf("Yest %s", modTime.Format("15:04"))
	}

	if modTime.Year() == now.Year() {
		return modTime.Format("Jan 2 15:04")
	}

	return modTime.Format("Jan 2 2006 15:04")
}

func sameDay(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func humanSize(size int64) string {
	kb := float64(size) / 1024
	mb := kb / 1024
	if roundToTenths(mb) >= 0.1 {
		return fmt.Sprintf("%.1f MB", mb)
	}

	if roundToTenths(kb) >= 0.1 {
		return fmt.Sprintf("%.1f KB", kb)
	}

	return fmt.Sprintf("%d B", size)
}

func roundToTenths(value float64) float64 {
	return math.Round(value*10) / 10
}

func truncateText(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	if width <= 3 {
		return string(runes[:width])
	}

	return string(runes[:width-3]) + "..."
}

func fuzzyScore(candidate string, query string) (int, bool) {
	candidate = strings.ToLower(candidate)
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return 0, false
	}

	idx := 0
	lastMatch := -1
	score := 0

	for _, r := range query {
		found := false
		for idx < len(candidate) {
			if rune(candidate[idx]) == r {
				score += idx
				if lastMatch != -1 {
					score += idx - lastMatch - 1
				}
				lastMatch = idx
				idx++
				found = true
				break
			}
			idx++
		}

		if !found {
			return 0, false
		}
	}

	score += len(candidate) - len(query)
	if strings.Contains(candidate, query) {
		score -= len(query)
	}

	return score, true
}
