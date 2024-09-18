package kademlia

import (
	"encoding/gob"
	"errors"
	"log"
	"os"
)

type Kademlia struct {
	Routes *RoutingTable
}

func NewKademliaNode() *Kademlia {
	kademlia_node := Kademlia{}
	new_id := NewRandomKademliaID()
	new_me_contact := NewContact(new_id, GetLocalIP())
	kademlia_node.Routes = NewRoutingTable(new_me_contact)
	err := DumpNodeInfo(os.Getenv("STATE_FILE"), &kademlia_node)
	if err != nil {
		log.Println("Could not save node info to file")
	}
	return &kademlia_node
}

func DumpNodeInfo(filename string, kademlia_node *Kademlia) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(kademlia_node)
	if err != nil {
		return err
	}

	return nil
}

func LoadNodeInfo(filename string) (*Kademlia, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var kademlia Kademlia
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&kademlia)
	if err != nil {
		return nil, err
	}

	return &kademlia, nil
}

func InitNode() *Kademlia {
	file_name := os.Getenv("STATE_FILE")
	if _, err := os.Stat(file_name); err == nil {
		node, err := LoadNodeInfo(file_name)
		if err != nil {
			log.Fatalln("Could not load node state!")
		}
		return node
	} else if errors.Is(err, os.ErrNotExist) {
		return NewKademliaNode()
	} else {
		log.Fatalln("Weird file bug, not sure what to do here")
		return nil
	}

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
