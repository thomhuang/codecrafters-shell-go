package main

import "strings"

func parseUserInput(input string) string {
	var res strings.Builder

	var currWord strings.Builder
	inQuotes := false
	for _, c := range input {
		switch {
		case c == '\'': // toggle whether we're within quotes or not
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes: // if we encounter a space and we're not within quotes, we treat it as a separator between the command and its arguments
			if currWord.Len() > 0 {
				res.WriteString(currWord.String())
				res.WriteRune(' ')
				currWord.Reset() // if we run into repeated spaces, we don't want to add extra spaces to the result
			}
		default:
			currWord.WriteRune(c)
		}
	}

	if currWord.Len() > 0 { // get last "part" of the input if it exists
		res.WriteString(currWord.String())
	}

	return res.String()
}
