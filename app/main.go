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
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		command = strings.TrimRight(command, "\r\n")
		if command == "exit" {
			break
		}

		fmt.Printf("%s: command not found\n", command)
	}
}
