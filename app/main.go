package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// builtins is a map of shell built-in commands for quick lookup
var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
}

var pathExecutables map[string]string

func main() {
	// get all executables from the PATH environment variable
	pathExecutables = getPathExecutables(os.Getenv("PATH"))

	for {
		fmt.Print("$ ")

		reader := bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		userInput = strings.TrimRight(userInput, "\r\n")
		parts := strings.SplitN(userInput, " ", 2)
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

func echoCommand(args string) {
	fmt.Println(args)
}

func typeCommand(args string) {
	if _, exists := builtins[args]; exists {
		fmt.Printf("%s is a shell builtin\n", args)
	} else if path, exists := pathExecutables[args]; exists {
		fmt.Printf("%s is %s\n", args, path)
	} else {
		fmt.Printf("%s: not found\n", args)
	}
}

func pwdCommand() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	fmt.Println(dir)
}

func getPathExecutables(path string) map[string]string {
	pathExecutables := make(map[string]string)

	// take in path environment variable and split it into directories
	paths := strings.SplitSeq(path, string(os.PathListSeparator))
	// for each path, walk through the directory for each file
	for path := range paths {
		if len(path) == 0 {
			continue
		}

		currDir, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		// for each file, check if it is executable and add it to the path executable map
		for _, file := range currDir {
			if file.IsDir() {
				continue
			}

			fullPath := filepath.Join(path, file.Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			if runtime.GOOS == "windows" && isExecutable(fullPath) { // windows doesn't use POSIX permission bits, so we have to check the file extension instead
				pathExecutables[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = strings.TrimSuffix(fullPath, filepath.Ext(fullPath))
			} else if info.Mode().Perm()&0111 != 0 { // check if file permissions indicate executable
				pathExecutables[file.Name()] = fullPath
			}
		}
	}

	return pathExecutables
}

func isExecutable(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	pathext := os.Getenv("PATHEXT")
	if pathext == "" {
		pathext = ".COM;.EXE;.BAT;.CMD"
	}

	return slices.Contains(strings.Split(strings.ToLower(pathext), ";"), ext)
}
