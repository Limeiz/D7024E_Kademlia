package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	body, err := io.ReadAll(resp.Body)
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

	case "put":
		if len(os.Args) < 3 {
			fmt.Println("Usage: put <data>")
			return
		}
		data := url.QueryEscape(os.Args[2]) // in case of special characters
		url_path := fmt.Sprintf("/put?data=%s", data)
		SendHTTPRequest(server_port, url_path)

	case "get":
		if len(os.Args) < 3 {
			fmt.Println("Usage: get <hash>")
			return
		}
		hash := url.QueryEscape(os.Args[2])
		url_path := fmt.Sprintf("/get?hash=%s", hash)
		SendHTTPRequest(server_port, url_path)

	case "show-id":
		url_path := fmt.Sprintf("/getid")
		SendHTTPRequest(server_port, url_path)

	case "show-routing-table":
		url_path := fmt.Sprintf("/show-routing-table")
		SendHTTPRequest(server_port, url_path)

	case "show-storage":
		url_path := fmt.Sprintf("/show-storage")
		SendHTTPRequest(server_port, url_path)

	case "exit":
		fmt.Println("Exiting node...")
		os.Exit(0)

	case "help":
		fmt.Println("RPC commands:")
		fmt.Println("ping <to>")
		fmt.Println("put <data>")
		fmt.Println("get <hash>")

		fmt.Println("Show commands:")
		fmt.Println("Usage: show-[node_variable]")
		fmt.Println("Availible commands:")
		fmt.Println("show-id")
		fmt.Println("show-routing-table")
		fmt.Println("show-storage")

		fmt.Println("exit")

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
