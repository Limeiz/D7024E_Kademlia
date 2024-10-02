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

func (network *Network) PutController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/put")
	data := request.FormValue("data")
	if data == "" {
		http.Error(response, "Missing 'data' parameter", http.StatusBadRequest)
		return
	}

	hash, err := network.Node.Store([]byte(data))
	if err != nil {
		fmt.Fprintf(response, "Error: Could not store data: %v", err)
		return
	}
	fmt.Fprintf(response, "Data succesfully stored with hash: %s", hash)
}

func (network *Network) GetController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/get")
	hash := request.FormValue("hash")
	if hash == "" {
		http.Error(response, "Missing 'hash' parameter", http.StatusBadRequest)
		return
	}

	data, nodes, err := network.Node.LookupData(hash)
	node := nodes[0]
	if err != nil {
		fmt.Fprintf(response, "Error: Could not find data: %v", err)
		return
	}
	fmt.Fprintf(response, "Data found on node %s with hash %s: %s", node.ID.String(), hash, data)
}

func (network *Network) ShowRoutingTableController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/show-routing-table")
	fmt.Fprintf(response, "Routing Table:\n%s", network.Node.Routes.String())
}

func (network *Network) ShowStorageController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/show-storage")
	responseString := "Stored data:\n"
	for key, value := range network.Node.Storage {
		responseString += fmt.Sprintf("Key: %s, Value: %s\n", key.String(), value)
	}
	fmt.Fprintf(response, "%s", responseString)
}
