package main

import (
	"fmt"
	"os"
	"strings"
)

const HOME_PATH = "~"

// builtins is the current set of commands  handled by the shell itself rather than
// by an external executable.
var builtins = map[string]bool{
	"exit":     true,
	"echo":     true,
	"type":     true,
	"pwd":      true,
	"cd":       true,
	"complete": true,
}

func echoCommand(args []string) (output, errorMessage string) {
	return strings.Join(args, " "), ""
}

func typeCommand(args []string) (output, errorMessage string) {
	arg := strings.Join(args, " ")

	if _, exists := builtins[arg]; exists {
		return fmt.Sprintf("%s is a shell builtin", arg), ""
	} else if path, exists := pathExecutables[arg]; exists {
		return fmt.Sprintf("%s is %s", arg, path), ""
	} else {
		return "", fmt.Sprintf("%s: not found", arg)
	}
}

func pwdCommand() (output, errorMessage string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Sprintf("Error getting current directory: %v", err)
	}

	return dir, ""
}

func cdCommand(args []string) (output, errorMessage string) {
	path := strings.Join(args, " ")
	if path == HOME_PATH {
		homeDir, _ := os.UserHomeDir()
		path = homeDir
	}

	err := os.Chdir(path)
	if err != nil {
		return "", fmt.Sprintf("cd: %s: No such file or directory", path)
	}

	return "", ""
}
