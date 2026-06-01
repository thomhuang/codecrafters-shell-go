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
		args, redirectInfo := extractRedirects(args)

		// Determine where stdout goes
		stdout := os.Stdout
		stderr := os.Stderr
		if redirectInfo != nil {
			switch redirectInfo.Type {
			case STDOUT:
				stdout = redirectInfo.File
			case STDERR:
				stderr = redirectInfo.File
			}
			defer redirectInfo.File.Close()
		}

		switch cmd {
		case "exit":
			return
		case "echo":
			output, errMsg := echoCommand(args)
			printResult(stdout, stderr, output, errMsg)
		case "type":
			output, errMsg := typeCommand(args)
			printResult(stdout, stderr, output, errMsg)
		case "pwd":
			output, errMsg := pwdCommand()
			printResult(stdout, stderr, output, errMsg)
		case "cd":
			_, errMsg := cdCommand(args)
			if errMsg != "" {
				printResult(stdout, stderr, "", errMsg)
			}
		default:
			if _, exists := pathExecutables[cmd]; exists {
				c := exec.Command(cmd, args...)
				c.Stdout = stdout
				c.Stderr = stderr
				c.Run()
			} else {
				printResult(stdout, stderr, "", fmt.Sprintf("%s: command not found", cmd))
			}
		}

		if redirectInfo != nil {
			redirectInfo.File.Close()
		}
	}
}

func printResult(stdout *os.File, stderr *os.File, output, errMsg string) {
	if errMsg != "" {
		fmt.Fprintln(stderr, errMsg)
	} else if output != "" {
		fmt.Fprintln(stdout, output)
	}
}
