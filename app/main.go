package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// builtins is a map of shell built-in commands for quick lookup
var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
	"pwd":  true,
	"cd":   true,
}

var pathExecutables map[string]string

func parseUserInput(input string) string {
	var res strings.Builder

	var currWord strings.Builder
	inQuotes := false
	for _, c := range input {
		switch {
		case c == '\'': // whenever we encounter a single quote, we toggle whether we're within quotes or not
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes: // if we encounter a space and we're not within quotes, we treat it as a separator between the command and its arguments
			if currWord.Len() > 0 {
				res.WriteString(currWord.String())
				res.WriteRune(' ')
				currWord.Reset()
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

func main() {
	reader := bufio.NewReader(os.Stdin)

	// get all executables from the PATH environment variable
	pathExecutables = getPathExecutables(os.Getenv("PATH"))
	for {
		fmt.Print("$ ")

		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		var parts []string

		userInput = parseUserInput(strings.TrimSpace(userInput))

		parts = strings.SplitN(userInput, " ", 2)
		cmd, args := parts[0], ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch cmd {
		case "exit":
			return
		case "echo":
			echoCommand(args)
		case "type":
			typeCommand(args)
		case "pwd":
			pwdCommand()
		case "cd":
			cdCommand(args)
		default:
			if _, exists := pathExecutables[cmd]; exists {
				cmd := exec.Command(cmd, strings.Split(args, " ")...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				fmt.Printf("%s: command not found\n", cmd)
			}
		}
	}
}
