package kademlia

type Kademlia struct {
	Routes *RoutingTable
}

func InitNode(premade_id ...*KademliaID) *Kademlia {
	kademlia_node := Kademlia{}
	var new_me_contact Contact
	if len(premade_id) > 0 {
		new_id := premade_id[0]
		new_me_contact = NewContact(new_id, GetLocalIP())
	} else {
		new_id := NewRandomKademliaID()
		new_me_contact = NewContact(new_id, GetLocalIP())
	}
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
