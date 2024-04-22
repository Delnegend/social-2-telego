package utils

import "strings"

func EscapeTelegramChar(s string) string {
	for _, char := range []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		s = strings.Replace(s, char, "\\"+char, -1)
	}
	return s
}
