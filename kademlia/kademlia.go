package kademlia

import (
	"os"
	"log"
	"fmt"
)

type Kademlia struct {
	Routes *RoutingTable
	Network *Network
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

func (kademlia *Kademlia) Ping(contact *Contact) (error){
	message, err := kademlia.Network.SendMessageAndWait(contact, PING, REQUEST, nil)
	if err != nil{
		log.Printf("Error: Ping could not be sent to %s\n", contact.Address)
		kademlia.Routes.RemoveContact(&message.Header.SenderID) // Since the receiver of this ping is now a sender of pong
		return fmt.Errorf("Error: Ping could not be sent to %s\n", contact.Address)
	}
	new_id := NewKademliaID(message.Header.SenderID.String()) // Since message will go out of scope
	contact.ID = new_id
	kademlia.Routes.AddContact(*contact)
	return nil
}

func (kademlia *Kademlia) Pong(contact *Contact, message_id *KademliaID){
	go kademlia.Network.SendMessage(contact, PING, RESPONSE, nil, message_id)
	kademlia.Routes.AddContact(*contact)
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
