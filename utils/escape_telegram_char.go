package utils

import "strings"

// add "ESCAPE_CHAR" in front of the telegram special characters
func EscapeTelegramChar(s string) string {
	for _, char := range []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		s = strings.Replace(s, char, "ESCAPE_CHAR"+char, -1)
	}
	return s
}
