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

func (network *Network) GetID(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/getid")

	fmt.Fprintf(response, "This node's id is: %s", network.Node.Routes.Me.ID.String())
}

func (network *Network) PingController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/ping")
	ping_address := request.FormValue("to")
	if ping_address == "" {
		http.Error(response, "Missing 'to' parameter", http.StatusBadRequest)
		return
	}

	new_contact := NewContact(NewRandomKademliaID(), ping_address)
	err := network.Node.Ping(&new_contact)
	if err != nil {
		fmt.Fprintf(response, "Error: <REQUEST,PING> cound not be sent: %v", err)
		return
	}
	fmt.Fprintf(response, "<RESPONSE,PING> recieved from: %s", ping_address)
}

func (network *Network) ShowRoutingTableController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/show-routing-table")
	fmt.Fprintf(response, "Routing Table:\n%s", network.Node.Routes.String())
}
