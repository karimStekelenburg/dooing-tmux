package model

import (
	"regexp"
	"sort"
)

var tagRegex = regexp.MustCompile(`#(\w+)`)

// extractTags returns all unique #tag values from a string in order of first occurrence.
func extractTags(text string) []string {
	matches := tagRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		tag := m[1]
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	return result
}

// GetAllTags returns a sorted, unique list of all tags across the given todos.
func GetAllTags(todos []*Todo) []string {
	seen := make(map[string]bool)
	for _, t := range todos {
		for _, tag := range t.ExtractAllTags() {
			seen[tag] = true
		}
	}
	result := make([]string, 0, len(seen))
	for tag := range seen {
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}
