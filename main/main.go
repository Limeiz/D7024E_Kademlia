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
	kademlia_node_state := kademlia.InitNode()
	if kademlia_node_state == nil {
		log.Fatalln("Could not create or load node state!")
		return
	}
	if len(os.Args) > 1 {
		cli.Init()
	} else {
		log.Println("Starting kademlia on node ", kademlia_node_state.Routes.Me.ID)
		comm_port, err := strconv.Atoi(os.Getenv("COMMUNICATION_PORT"))

		if err != nil {
			log.Fatalln("Could not parse the COMMUNICATION_PORT env variable")
		}
		kademlia.OpenPortAndListen(comm_port)
	}
}
