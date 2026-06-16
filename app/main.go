package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// executables maps an executable's command name to its full path on disk,
// built once at startup from the PATH environment variable.
var executables map[string]string

func main() {
	executables = getPathExecutables(os.Getenv("PATH"))
	buildCmdCompletionTrie()

	// Our own raw-mode line editor (replaces the readline package): it reads
	// input one key at a time so we can handle tab-completion, backspace, and
	// Ctrl-C/Ctrl-D ourselves.
	editor := NewLineEditor("$ ")

	for {
		userInput, err := editor.ReadLine()
		if err == ErrInterrupt || err == io.EOF { // Ctrl-C or Ctrl-D: exit the shell
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
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
			switch redirectInfo.typ {
			case STDOUT:
				stdout = redirectInfo.file
			case STDERR:
				stderr = redirectInfo.file
			}
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
			output, errMsg := cdCommand(args)
			printResult(stdout, stderr, output, errMsg)
		case "complete":
			output, errMsg := completeCommand(args)
			printResult(stdout, stderr, output, errMsg)
		default:
			if _, exists := executables[cmd]; exists {
				c := exec.Command(cmd, args...)
				c.Stdout = stdout
				c.Stderr = stderr
				c.Run()
			} else {
				printResult(stdout, stderr, "", fmt.Sprintf("%s: command not found", cmd))
			}
		}

		if redirectInfo != nil {
			redirectInfo.file.Close()
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
