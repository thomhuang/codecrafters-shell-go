package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
}

func main() {
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
			} else {
				fmt.Printf("%s: not found\n", args)
			}
		default:
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
