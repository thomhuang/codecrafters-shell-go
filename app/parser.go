package main

import "strings"

// doubleQuoteEscapable lists the characters whose backslash-escape is honored
// inside double quotes; bash leaves "\x" untouched for any other x.
var doubleQuoteEscapable = map[rune]bool{
	'\\': true,
	'"':  true,
	'$':  true,
	'`':  true,
	'\n': true,
}

const BACKSLASH = '\\'
const SINGLE_QUOTE = '\''
const DOUBLE_QUOTE = '"'

// parseUserInput tokenizes a line of shell input, honoring single quotes,
// double quotes, and backslash escapes.
func parseUserInput(input string) []string {
	var tokens []string
	var current strings.Builder
	quoteChar := rune(0) // 0 means we're not inside quotes

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		c := runes[i]

		switch {
		case c == BACKSLASH && quoteChar != SINGLE_QUOTE: // backslash is only special outside single quotes
			i += writeEscaped(&current, runes, i, quoteChar)
		case quoteChar == 0 && (c == DOUBLE_QUOTE || c == SINGLE_QUOTE): // opening quote
			quoteChar = c // write the opening quote char
		case c == quoteChar: // closing quote
			quoteChar = 0 // reset quote state
		case c == ' ' && quoteChar == 0: // unquoted space separates tokens
			tokens = flushToken(tokens, &current)
		default:
			current.WriteRune(c)
		}
	}

	return flushToken(tokens, &current)
}

// flushToken appends the accumulated token (if any) to tokens and resets the
// builder, returning the updated slice.
func flushToken(tokens []string, current *strings.Builder) []string {
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
		current.Reset()
	}
	return tokens
}

// writeEscaped handles a backslash at runes[i] and returns the number of extra
// runes consumed (0 for a trailing backslash, 1 otherwise).
func writeEscaped(current *strings.Builder, runes []rune, i int, quoteChar rune) int {
	if i+1 >= len(runes) {
		current.WriteRune(BACKSLASH) // trailing backslash stays literal
		return 0
	}

	next := runes[i+1]
	switch quoteChar {
	case 0: // outside quotes, backslash escapes any character
		current.WriteRune(next)
	case DOUBLE_QUOTE: // inside double quotes, only a select few chars are escapable
		if doubleQuoteEscapable[next] {
			current.WriteRune(next)
		} else {
			current.WriteRune(BACKSLASH)
			current.WriteRune(next)
		}
	default:
		current.WriteRune(BACKSLASH)
		current.WriteRune(next)
	}
	return 1
}
