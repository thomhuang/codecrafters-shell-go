package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/chzyer/readline"
)

func main() {
	pathExecutables = getPathExecutables(os.Getenv("PATH"))
	autoCompleteTrie = buildCompletionTrie()

	// use readline package to make our shell be in raw mode, which allows us to read user input one key at a time and handle special keys like tab-autocomplete, backspace and Ctrl-C.
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    &commandAutoComplete{trie: autoCompleteTrie},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		userInput, err := rl.Readline()
		if err == readline.ErrInterrupt { // Ctrl-C: drop the line, new prompt
			continue
		} else if err == io.EOF { // Ctrl-D
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
