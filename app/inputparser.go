package main

import "strings"

func parseUserInput(input string) []string {
	var parts []string

	var currWord strings.Builder
	inQuotes := false
	for _, c := range input {
		switch {
		case c == '\'': // toggle whether we're within quotes or not
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes: // if we encounter a space and we're not within quotes, we treat it as a separator between the command and its arguments
			if currWord.Len() > 0 {
				parts = append(parts, currWord.String())
				currWord.Reset()
			}
		default:
			currWord.WriteRune(c)
		}
	}

	if currWord.Len() > 0 {
		parts = append(parts, currWord.String())
	}

	return parts
}
