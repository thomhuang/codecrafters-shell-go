package main

import "strings"

func parseUserInput(input string) []string {
	var parts []string

	var currWord strings.Builder
	quoteChar := rune(0) // 0 means not in quotes
	for i := 0; i < len(input); {
		c := rune(input[i])

		switch {
		case quoteChar == 0 && c == '\\': // if we encounter a backslash, we skip the next character and add it to the current word
			i++
			if i < len(input) {
				c = rune(input[i])
				currWord.WriteRune(c)
			}
		case quoteChar == 0 && (c == '"' || c == '\''): // whenever we encounter a double quote, we toggle whether we're within quotes or not
			quoteChar = c
		case c == quoteChar:
			quoteChar = 0
		case c == ' ' && quoteChar == 0:
			if currWord.Len() > 0 {
				parts = append(parts, currWord.String())
				currWord.Reset()
			}
		default:
			currWord.WriteRune(c)
		}

		i++
	}

	if currWord.Len() > 0 {
		parts = append(parts, currWord.String())
	}

	return parts
}
