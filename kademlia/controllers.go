package kademlia

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func BeginResponse(request *http.Request, path string) {
	clientIP := request.RemoteAddr
	if forwarded := request.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}
	log.Printf("Got request to path \"%s\" from %s", path, clientIP)
}

func (network *Network) DefaultController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/")

	io.WriteString(response, "Try some of these commands:\n")
	io.WriteString(response, "/ping?to=<address>\n")
}

func (network *Network) PingController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/ping")
	ping_address := request.FormValue("to")
	if ping_address == "" {
		http.Error(response, "Missing 'to' parameter", http.StatusBadRequest)
		return
	}

	_, err := network.SendMessageAndWaitByIP(ping_address, PING, REQUEST, nil)
	if err != nil {
		fmt.Fprintf(response, "Error: Ping cound not be sent: %v", err)
		return
	}
	fmt.Fprintf(response, "Pong recieved from: %s", ping_address)
}
