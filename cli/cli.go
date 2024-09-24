package cli

import (
	"fmt"
	"net/http"
	"os"
	"io/ioutil"
)

func Init(server_port int) {
	if len(os.Args) < 2 {
		fmt.Println("No command provided. Use 'help' for available commands.")
		return
	}

	switch os.Args[1] {
	case "ping":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ping <to>")
			return
		}
		ip := os.Args[2]
		url := fmt.Sprintf("http://localhost:%d/ping?to=%s", server_port, ip)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to send ping: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response: %v\n", err)
			return
		}
		fmt.Printf("Response: %s\n", body)
		break

	case "help":
		fmt.Println("Commands:")
		fmt.Println("ping <to>")
		break

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
