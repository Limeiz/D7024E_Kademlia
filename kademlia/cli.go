package kademlia

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//https://www.mizouzie.dev/articles/3-ways-to-read-input-with-go-cli/

func StartCLI(kademlia *Kademlia) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter command")
	fmt.Print("> ")

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		command := strings.Fields(input)

		if len(command) == 0 {
			fmt.Println("Please enter a command.")
			fmt.Print("> ")

			continue
		}

		switch command[0] {
		case "put":
			if len(command) < 2 || len(command) > 2 {
				fmt.Println("Error: 'put' command requires an argument.")
				fmt.Println("Usage: put <file>")
				fmt.Print("> ")
				continue
			}

			value := command[1]

			hash := kademlia.Store(value)
			fmt.Printf("File uploaded successfully. Hash: %s\n", hash)

			//set limit to string to 255bytes

		case "get":
			if len(command) < 2 || len(command) > 2 {
				fmt.Println("Error: 'get' command requires an argument.")
				fmt.Println("Usage: get <hash>")
				fmt.Print("> ")
				continue
			}

			hash := command[1]
			data, node := kademlia.LookupData(hash)
			if data != nil {
				fmt.Printf("Data retrieved from node %s: \n%s\n", node, string(data))
			} else {
				fmt.Println("Error: Could not retrieve data for the given hash.")
			}

		case "exit":
			fmt.Println("Exiting...")
			os.Exit(0)

		default:
			fmt.Println("Unknown command. Use 'put <file>', 'get <hash>', or 'exit'.")
			fmt.Print("> ")
		}
	}
}
