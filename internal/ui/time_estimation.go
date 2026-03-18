package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// parseTimeEstimation parses a time string like "15m", "2h", "1d", "0.5w"
// into hours (float64). Returns (hours, true) on success, (0, false) on error.
func parseTimeEstimation(input string) (float64, bool) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return 0, false
	}

	unit := input[len(input)-1:]
	numStr := input[:len(input)-1]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil || val <= 0 {
		return 0, false
	}

	switch strings.ToLower(unit) {
	case "m":
		return val / 60.0, true
	case "h":
		return val, true
	case "d":
		return val * 8.0, true
	case "w":
		return val * 40.0, true
	}
	return 0, false
}

// formatTimeEstimation converts hours back to a best-fit human-readable string.
// < 1h   → minutes (e.g., "30m")
// < 8h   → hours   (e.g., "2h")
// < 40h  → days    (e.g., "1d")
// >= 40h → weeks   (e.g., "0.5w")
func formatTimeEstimation(hours float64) string {
	if hours <= 0 {
		return ""
	}

	switch {
	case hours < 1.0:
		mins := hours * 60.0
		return fmt.Sprintf("[≈ %sm]", formatFloat(mins))
	case hours < 8.0:
		return fmt.Sprintf("[≈ %sh]", formatFloat(hours))
	case hours < 40.0:
		days := hours / 8.0
		return fmt.Sprintf("[≈ %sd]", formatFloat(days))
	default:
		weeks := hours / 40.0
		return fmt.Sprintf("[≈ %sw]", formatFloat(weeks))
	}
}

// formatFloat renders a float as an integer if it's a whole number,
// otherwise with minimal decimal places.
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	// Use up to 2 decimal places, stripping trailing zeros.
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
