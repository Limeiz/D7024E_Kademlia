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
		return
	}
	if len(os.Args) > 1 {
		cli.Init(server_port)
	} else {
		var kademlia_node_state *kademlia.Kademlia
		bootstrap_id_string := os.Getenv("BOOTSTRAP_ID")
		bootstrap_node_id := kademlia.NewKademliaID(bootstrap_id_string)
		kademlia_node_state = kademlia.InitNode(bootstrap_node_id)

		server_timeout, timeout_err := strconv.Atoi(os.Getenv("NETWORK_TIMEOUT"))

		if timeout_err != nil{
			log.Fatalln("Could not parse the NETWORK_TIMEOUT env variable")
			return
		}

		net := kademlia.InitNetwork(kademlia_node_state, server_timeout)
		kademlia_node_state.Network = net
		log.Println("Starting kademlia on node ", kademlia_node_state.Routes.Me.ID)
		comm_port, comm_err := strconv.Atoi(os.Getenv("COMMUNICATION_PORT"))

		if comm_err != nil {
			log.Fatalln("Could not parse the COMMUNICATION_PORT env variable")
			return
		}

		go net.OpenPortAndListen(comm_port)
		if os.Getenv("NODE_TYPE") != "bootstrap" {
			go kademlia_node_state.InitNetwork(&kademlia_node_state.Routes.Me)
			// kademlia_node_state.RefreshBuckets(kademlia_node_state.Routes.Me.ID) // have to do this after net is set up
		}
		net.ServerInit()
		net.ServerStart(server_port)
	}
}
