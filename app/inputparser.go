package main

import "strings"

var specialCharacters = map[rune]bool{
	'\\': true,
	'"':  true,
	'\'': true,
	'$':  true,
	'`':  true,
	'\n': true,
}

const SINGLE_QUOTE = '\''
const DOUBLE_QUOTE = '"'

func parseUserInput(input string) []string {
	var parts []string

	var currPart strings.Builder
	quoteChar := rune(0) // 0 means not in quotes
	for i := 0; i < len(input); {
		c := rune(input[i])

		switch {
		case c == '\\' && quoteChar != SINGLE_QUOTE: // backslash is only special outside single quotes
			if i+1 < len(input) {
				next := rune(input[i+1])
				switch quoteChar {
				case DOUBLE_QUOTE:
					if specialCharacters[next] {
						currPart.WriteRune(next)
					} else {
						currPart.WriteRune(c)
						currPart.WriteRune(next)
					}
				case 0:
					currPart.WriteRune(next)
				default:
					currPart.WriteRune(c)
					currPart.WriteRune(next)
				}
				i++ // skip the escaped character
			} else {
				currPart.WriteRune(c)
			}
		case quoteChar == 0 && (c == DOUBLE_QUOTE || c == SINGLE_QUOTE): // whenever we encounter a double quote, we toggle whether we're within quotes or not
			quoteChar = c
		case c == quoteChar:
			quoteChar = 0
		case c == ' ' && quoteChar == 0:
			if currPart.Len() > 0 {
				parts = append(parts, currPart.String())
				currPart.Reset()
			}
		default:
			currPart.WriteRune(c)
		}

		i++
	}

	if currPart.Len() > 0 {
		parts = append(parts, currPart.String())
	}

	return parts
}
