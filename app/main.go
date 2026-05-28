package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
}

func main() {
	pathExecutables := getPathExecutables(os.Getenv("PATH"))
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
			fmt.Println(args)
		case "type":
			if _, exists := builtins[args]; exists {
				fmt.Printf("%s is a shell builtin\n", args)
			} else if path, exists := pathExecutables[args]; exists {
				fmt.Printf("%s is %s\n", args, path)
			} else {
				fmt.Printf("%s: not found\n", args)
			}
		default:
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}

func getPathExecutables(path string) map[string]string {
	var pathExecutables map[string]string

	paths := strings.SplitSeq(path, string(os.PathListSeparator))
	for path := range paths {
		if len(path) == 0 {
			continue
		}

		currDir, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, file := range currDir {
			if file.IsDir() {
				continue
			}

			fullPath := filepath.Join(path, file.Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			if info.Mode().Perm()&0111 != 0 {
				pathExecutables[file.Name()] = path
			}
		}
	}

	return pathExecutables
}
