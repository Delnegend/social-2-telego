package utils

import "strings"

// Escape special characters in a string using a custom escape character
func EscapeSpecialChars(s string, escapeChar string) string {
	for _, char := range []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		s = strings.Replace(s, char, escapeChar+char, -1)
	}
	return s
}
