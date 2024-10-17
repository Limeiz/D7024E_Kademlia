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

	fmt.Printf("Received data: %s\n", data)

	if data == "" {
		http.Error(response, "Missing 'data' parameter", http.StatusBadRequest)
		return
	}

	hash, err := network.Node.Store([]byte(data))

	if err != nil {
		fmt.Fprintf(response, "Error: Could not store data: %v", err)
		return
	}

	fmt.Fprintf(response, "Data successfully stored with hash: %s", hash)
	fmt.Printf("Data stored with hash: %s\n", hash)
}

func (network *Network) GetController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/get")
	hash := request.FormValue("hash")

	fmt.Printf("Received hash: %s\n", hash)

	if hash == "" {
		http.Error(response, "Missing 'hash' parameter", http.StatusBadRequest)
		return
	}

	data, nodes, err := network.Node.LookupData(hash)
	if err != nil {
		fmt.Fprintf(response, "Error: Could not perform lookup: %v", err)
		return
	}

	log.Printf("Get controller found data %s with length %d", data, len(data))

	if data != "" {
		if len(nodes) > 0 {
			fmt.Fprintf(response, "Data found for hash %s: %s on node %v", hash, data, nodes[0].ID.String())
		} else {
			fmt.Fprintf(response, "Data found for hash %s: %s, on unknown node", hash, data) // should not happen
		}
		return
	}

	if len(nodes) == 0 {
		fmt.Fprintf(response, "Data not found for hash %s, and no closer nodes are available.", hash)
		return
	}

	fmt.Fprintf(response, "Data not found for hash %s. Closest nodes are:\n", hash)
	// for _, node := range nodes {
	// 	fmt.Fprintf(response, "- Node ID: %s, Address: %s\n", node.ID.String(), node.Address)
	// }
}

func (network *Network) ExitController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/exit")
	network.Node.Quit()
	fmt.Fprintf(response, "Shutting down node...")
}

func (network *Network) ShowRoutingTableController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/show-routing-table")
	fmt.Fprintf(response, "Routing Table:\n%s", network.Node.Routes.String())
}

func (network *Network) ShowStorageController(response http.ResponseWriter, request *http.Request) {
	BeginResponse(request, "/show-storage")
	responseString := "Stored data:\n"
	for key, value := range network.Node.Storage {
		responseString += fmt.Sprintf("Key: %s, Value: %s\n", key.String(), value.Value)
	}
	fmt.Fprintf(response, "%s", responseString)
}
