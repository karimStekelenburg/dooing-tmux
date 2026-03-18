package ui

import "regexp"

// tagRegexForUI matches #word patterns for display highlighting.
var tagRegexForUI = regexp.MustCompile(`#\w+`)
