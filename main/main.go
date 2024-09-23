// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/cli"
	"d7024e/kademlia"
	"log"
	"os"
	"strconv"
)

func main() {
	server_port, server_err := strconv.Atoi(os.Getenv("SERVER_PORT"))

	if server_err != nil {
		log.Fatalln("Could not parse the SERVER_PORT env variable")
	}
	if len(os.Args) > 1 {
		cli.Init(server_port)
	} else {
		kademlia_node_state := kademlia.InitNode()
		log.Println("Starting kademlia on node ", kademlia_node_state.Routes.Me.ID)
		comm_port, comm_err := strconv.Atoi(os.Getenv("COMMUNICATION_PORT"))

		if comm_err != nil {
			log.Fatalln("Could not parse the COMMUNICATION_PORT env variable")
		}

		go kademlia.OpenPortAndListen(comm_port)
		kademlia.ServerInit()
		kademlia.ServerStart(server_port)
	}
}
