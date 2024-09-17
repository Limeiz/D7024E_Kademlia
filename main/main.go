// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/kademlia"
	"log"
)

func main() {
	id := kademlia.NewRandomKademliaID()
	log.Println("Starting kademlia on node %d", id)

	kademlia.OpenPortAndListen(8080)
	for {
	}
}
