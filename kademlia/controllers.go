package kademlia

import (
	"log"
	"net/http"
	"io"
	"fmt"
)

func BeginResponse(request *http.Request, path string) {
	clientIP := request.RemoteAddr
	if forwarded := request.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}
	log.Printf("Got request to path \"%s\" from %s", path, clientIP)
}

func DefaultController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/")

	io.WriteString(response, "Try some of these commands:\n")
	io.WriteString(response, "/ping?to=<address>\n")
}

func PingController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/ping")
	ping_address := request.FormValue("to")
	if ping_address == "" {
		http.Error(response, "Missing 'to' parameter", http.StatusBadRequest)
		return
	}

	SendPingMessageByIP(ping_address)
	fmt.Fprintf(response, "Ping sent to: %s", ping_address)
}