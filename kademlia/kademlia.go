package kademlia

import (
	"os"
)

type Kademlia struct {
	Routes *RoutingTable
}

func InitNode(bootstrap_id *KademliaID) *Kademlia {
	kademlia_node := Kademlia{}
	bootstrap_contact := NewContact(bootstrap_id, os.Getenv("BOOTSTRAP_NODE"))
	if os.Getenv("NODE_TYPE") == "bootstrap"{
		kademlia_node.Routes = NewRoutingTable(bootstrap_contact)
	}else{
		new_id := NewRandomKademliaID()
		new_me_contact := NewContact(new_id, GetLocalIP())
		kademlia_node.Routes = NewRoutingTable(new_me_contact)
		kademlia_node.Routes.AddContact(bootstrap_contact)
	}
	return &kademlia_node
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
