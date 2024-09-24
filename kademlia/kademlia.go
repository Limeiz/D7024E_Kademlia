package kademlia

import (
)

type Kademlia struct {
	Routes *RoutingTable
}

func InitNode() *Kademlia {
	kademlia_node := Kademlia{}
	new_id := NewRandomKademliaID()
	new_me_contact := NewContact(new_id, GetLocalIP())
	kademlia_node.Routes = NewRoutingTable(new_me_contact)
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
