package main

import "strings"

func parseUserInput(input string) []string {
	var parts []string

	var currWord strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	for _, c := range input {
		switch {
		case c == '"': // whenever we encounter a double quote, we toggle whether we're within quotes or not
			inDoubleQuotes = !inDoubleQuotes
		case c == '\'' && !inDoubleQuotes: // toggle whether we're within quotes or not
			inSingleQuotes = !inSingleQuotes
		case c == ' ' && !inSingleQuotes && !inDoubleQuotes: // if we encounter a space and we're not within quotes, we treat it as a separator between the command and its arguments
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
