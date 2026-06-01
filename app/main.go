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

		parts = parseUserInput(strings.TrimSpace(userInput))

		if len(parts) == 0 {
			continue
		}
		cmd := parts[0]
		args := parts[1:]

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
				cmd := exec.Command(cmd, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				fmt.Printf("%s: command not found\n", cmd)
			}
		}
	}
}
