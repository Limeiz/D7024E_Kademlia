package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func SendHTTPRequest(server_port int, http_path string) {
	url := fmt.Sprintf("http://localhost:%d%s", server_port, http_path)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n", body)
}

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
		url_path := fmt.Sprintf("/ping?to=%s", ip)

		SendHTTPRequest(server_port, url_path)
		break

	case "show-id":
		url_path := fmt.Sprintf("/getid")
		SendHTTPRequest(server_port, url_path)
		break

	case "show-routing-table":
		url_path := fmt.Sprintf("/show-routing-table")
		SendHTTPRequest(server_port, url_path)
		break

	case "help":
		fmt.Println("RPC commands:")
		fmt.Println("ping <to>")

		fmt.Println("Show commands:")
		fmt.Println("Usage: show-[node_variable]")
		fmt.Println("Availible commands:")
		fmt.Println("show-id")
		fmt.Println("show-routing-table")
		break

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
