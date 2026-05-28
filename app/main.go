package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Print("$ ")

		reader := bufio.NewReader(os.Stdin)
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		parts := strings.SplitN(strings.TrimRight(userInput, "\r\n"), " ", 2)
		cmd, args := parts[0], ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch cmd {
		case "exit":
			return
		case "echo":
			fmt.Println(args)
		default:
			fmt.Printf("%s: command not found\n", cmd)
		}
	}
}
