package main

import (
	"fmt"
	"os"
)

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

func cdCommand(path string) {
	if path == "~" {
		homeDir, _ := os.UserHomeDir()
		path = homeDir
	}

	err := os.Chdir(path)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", path)
	}
}
