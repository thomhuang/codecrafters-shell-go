package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
	"pwd":  true,
	"cd":   true,
}

var pathExecutables map[string]string

func main() {
	reader := bufio.NewReader(os.Stdin)
	pathExecutables = getPathExecutables(os.Getenv("PATH"))

	for {
		fmt.Print("$ ")

		userInput, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		parts := parseUserInput(strings.TrimSpace(userInput))
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		// Extract redirect operators from args
		args, stdoutRedirect := extractRedirects(args)

		// Determine where stdout goes
		stdout := os.Stdout
		if stdoutRedirect != nil {
			stdout = stdoutRedirect
			defer stdoutRedirect.Close()
		}

		switch cmd {
		case "exit":
			return
		case "echo":
			output, errMsg := echoCommand(args)
			printResult(stdout, output, errMsg)
		case "type":
			output, errMsg := typeCommand(args)
			printResult(stdout, output, errMsg)
		case "pwd":
			output, errMsg := pwdCommand()
			printResult(stdout, output, errMsg)
		case "cd":
			_, errMsg := cdCommand(args)
			if errMsg != "" {
				fmt.Fprintln(os.Stderr, errMsg)
			}
		default:
			if _, exists := pathExecutables[cmd]; exists {
				c := exec.Command(cmd, args...)
				c.Stdout = stdout
				c.Stderr = os.Stderr
				c.Run()
			} else {
				fmt.Fprintf(os.Stderr, "%s: command not found\n", cmd)
			}
		}

		if stdoutRedirect != nil {
			stdoutRedirect.Close()
		}
	}
}

func printResult(stdout *os.File, output, errMsg string) {
	if errMsg != "" {
		fmt.Fprintln(os.Stderr, errMsg)
	} else if output != "" {
		fmt.Fprintln(stdout, output)
	}
}
