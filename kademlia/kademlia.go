package kademlia

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

const alpha = 3

type Kademlia struct {
	Routes  *RoutingTable
	Network *Network
	Storage map[KademliaID]string
}

type FindValueResponse struct {
	Value           string
	ClosestContacts []Contact // The list of closer contacts, if no value is found
}

type ContactResponse struct { // not needed anymore, redo
	contacts []Contact
	err      error
}

type StoreData struct {
	Key   KademliaID
	Value string
}

// Please use these instead of creating a seriliaze and deserialize for every type of struct
func SerializeData[T any](data T) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DeserializeData[T any](data []byte) (T, error) {
	var result T
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	// Decode the data into the result
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func InitNode(bootstrap_id *KademliaID) *Kademlia {
	kademlia_node := Kademlia{
		Storage: make(map[KademliaID]string),
	}
	bootstrap_contact := NewContact(bootstrap_id, os.Getenv("BOOTSTRAP_NODE"))
	if os.Getenv("NODE_TYPE") == "bootstrap" {
		kademlia_node.Routes = NewRoutingTable(bootstrap_contact)
	} else {
		new_id := NewRandomKademliaID()
		new_me_contact := NewContact(new_id, GetLocalIP())
		kademlia_node.Routes = NewRoutingTable(new_me_contact)
		kademlia_node.Routes.AddContact(bootstrap_contact)
		// kademlia_node.RefreshBuckets(new_id) // look up ourself to fill Routes (step 3 and 4 in join)
	}

	return &kademlia_node
}

func (kademlia *Kademlia) Ping(contact *Contact) error {
	message, err := kademlia.Network.SendMessageAndWait(contact, PING, REQUEST, nil)
	if err != nil {
		log.Printf("Error: Ping could not be sent to %s\n", contact.Address)
		kademlia.Routes.RemoveContact(&message.Header.SenderID) // Since the receiver of this ping is now a sender of pong
		return fmt.Errorf("Error: Ping could not be sent to %s\n", contact.Address)
	}
	new_id := NewKademliaID(message.Header.SenderID.String()) // Since message will go out of scope
	contact.ID = new_id
	kademlia.Routes.AddContact(*contact)
	return nil
}

func (kademlia *Kademlia) Pong(contact *Contact, message_id *KademliaID) {
	go kademlia.Network.SendMessage(contact, PING, RESPONSE, nil, message_id)
	kademlia.Routes.AddContact(*contact)
}

func (kademlia *Kademlia) SendStoreRPC(contact *Contact, key *KademliaID, data string) error {
	store_message := StoreData{
		Key:   *key,
		Value: data,
	}

	serialized_data, data_err := SerializeData(store_message)
	if data_err != nil {
		log.Printf("Error: Could not serialize store data!")
		return fmt.Errorf("Error: Could not serialize store data!")
	}
	_, err := kademlia.Network.SendMessageAndWait(contact, STORE, REQUEST, serialized_data)
	if err != nil {
		log.Printf("Error: Store could not be sent to %s\n", contact.Address)
		return fmt.Errorf("Error: Store could not be sent to %s\n", contact.Address)
	}

	return nil

}

func (kademlia *Kademlia) RecieveStoreRPC(data *[]byte) {
	deserialized_data, err := DeserializeData[StoreData](*data)
	if err != nil {
		log.Printf("Error: Could not deserialize store data")
	}
	kademlia.Storage[deserialized_data.Key] = deserialized_data.Value
}

func (kademlia *Kademlia) LookupContact(target *Contact) []Contact { // iterativeFindNode
	// If contact doesn't respond, remove from routingTable

	kclosestContacts := kademlia.Routes.FindClosestContacts(target.ID, alpha)
	if len(kclosestContacts) == 0 {
		log.Println("No contacts found in the shortlist.")
		return nil
	}

	responseChan := make(chan ContactResponse, alpha)
	doneChan := make(chan struct{})
	activeRPCs := 0
	candidates := &ContactCandidates{
		contacts: kclosestContacts,
	}
	closestNode := candidates.contacts[0]

	visitedNodes := make(map[KademliaID]bool)

	for len(candidates.contacts) < IDLength {
		alphaContacts := []Contact{} // the nodes that we will send RPCs to
		for _, contact := range candidates.contacts {
			if !visitedNodes[*contact.ID] && len(alphaContacts) < alpha {
				alphaContacts = append(alphaContacts, contact)
				visitedNodes[*contact.ID] = true
				if contact.Less(&closestNode) {
					closestNode = contact
				}
			}
		}

		for _, contact := range alphaContacts {
			// Send the first batch of alpha parallel RPCs
			activeRPCs++
			go func(c Contact) {
				response, err := kademlia.SendFindNodeRPC(&c, target)
				responseChan <- ContactResponse{contacts: response, err: err}
			}(contact)
		}

		go processResponses(kademlia, responseChan, &candidates.contacts, visitedNodes, doneChan, closestNode, activeRPCs, *target) //ugly

		<-doneChan
	}

	candidates.Sort() // the canditades should be modified in processResponses
	return candidates.GetContacts(IDLength)
}

func (kademlia *Kademlia) SendFindNodeRPC(contact *Contact, target *Contact) ([]Contact, error) {
	serializedTarget, err := SerializeSingleContact(*target)
	if err != nil {
		log.Printf("Failed to serialize target contact: %v", err)
		return nil, err
	}

	data, err := kademlia.Network.SendMessageAndWait(contact, FIND_NODE, REQUEST, serializedTarget)
	if err != nil {
		log.Printf("Failed to send FIND_NODE RPC: %v", err)
		return nil, err
	}

	contacts, err := DeserializeContacts(data.Data)
	if err != nil {
		log.Printf("Failed to deserialize contacts: %v", err)
		return nil, err
	}

	return contacts, nil
}

func processResponses(kademlia *Kademlia, responseChan chan ContactResponse, contactCandidates *[]Contact, visitedNodes map[KademliaID]bool, doneChan chan struct{}, closestNode Contact, activeRPCs int, target Contact) {
	for len(*contactCandidates) < IDLength && activeRPCs > 0 {
		select {
		case response := <-responseChan:
			activeRPCs--
			if response.err != nil {
				log.Printf("Error in response: %v", response.err)
				continue // go to next response
			}

			for _, contact := range response.contacts {
				newContact := NewContact(contact.ID, contact.Address)
				newContact.CalcDistance(target.ID)
				kademlia.Routes.AddContact(newContact)

				if len(*contactCandidates) < IDLength {
					*contactCandidates = append(*contactCandidates, newContact)
				}

				if newContact.Less(&closestNode) {
					closestNode = newContact
				}

				// If we haven't visited this contact, send another RPC
				if !visitedNodes[*newContact.ID] {
					visitedNodes[*newContact.ID] = true
					activeRPCs++
					go func(c Contact) {
						response, err := kademlia.SendFindNodeRPC(&c, &target)
						responseChan <- ContactResponse{contacts: response, err: err}
					}(contact)
				}
			}

		default:
			// Search is finished
			if activeRPCs == 0 {
				doneChan <- struct{}{}
				return
			}
		}
	}
}

func (kademlia *Kademlia) ProcessFindContactMessage(data *[]byte, sender Contact) ([]byte, error) {
	target, err := DeserializeSingleContact(*data) // assumes msg data only holds target contact
	if err != nil {
		log.Printf("Failed to deserialize target contact: %v", err)
		return nil, err
	}
	targetID := target.ID

	closestContacts := kademlia.Routes.FindClosestContacts(targetID, IDLength)

	kademlia.Routes.AddContact(sender)

	// Serialize the list of closest contacts
	responseBytes, err := SerializeContacts(closestContacts)
	if err != nil {
		log.Printf("Failed to serialize contacts: %v", err)
		return nil, err
	}

	log.Printf("Sent closest contacts: ", responseBytes)
	return responseBytes, err
}

func (kademlia *Kademlia) RefreshBuckets(targetID *KademliaID) {
	// kademlia.LookupContact(&kademlia.Routes.Me)

	closestContacts := kademlia.Routes.FindClosestContacts(targetID, IDLength)
	if len(closestContacts) == 0 {
		log.Println("No contacts found for the target node")
		return
	}

	closestNeighbor := closestContacts[0]
	closestNeighborBucketIndex := kademlia.Routes.getBucketIndex(closestNeighbor.ID)

	// log.Println("should be bootstrap", closestContacts[0].ID)

	// Iterate over buckets further away than the closest neighbor
	for i := closestNeighborBucketIndex + 1; i < len(kademlia.Routes.buckets); i++ {
		bucket := kademlia.Routes.buckets[i]

		for contactElement := bucket.list.Front(); contactElement != nil; contactElement = contactElement.Next() {
			contact := contactElement.Value.(Contact)

			targetContact := Contact{
				ID:      targetID,
				Address: contact.Address,
			}

			// Send a ping to the nodes in this bucket, try find_node and iterative if doesn't work, have never come this far
			log.Printf("Refreshing bucket %d by pinging contact %s\n", i, contact.Address)
			kademlia.Network.Node.Ping(&targetContact)
		}
	}
}

func (kademlia *Kademlia) LookupData(hash string) (string, []Contact) { // iterativeFindValue
	var value string
	candidates := &ContactCandidates{}
	initialContacts := kademlia.Routes.FindClosestContacts(NewKademliaID(hash), alpha)
	visitedNodes := make(map[KademliaID]bool)

	responseChan := make(chan *FindValueResponse, alpha)
	doneChan := make(chan struct{})
	activeRPCs := 0

	for _, contact := range initialContacts {
		candidates.Append([]Contact{contact})
		visitedNodes[*contact.ID] = true
	}

	// Helper function to send FIND_VALUE requests
	goSendFindValueRPC := func(contact Contact) {
		activeRPCs++
		go func(c Contact) {
			defer func() { activeRPCs-- }()
			response, err := kademlia.SendFindValueRPC(&c, NewKademliaID(hash)) // CHANGE HERE to send and wait
			if err != nil {
				log.Printf("Failed to send SendFindValueRPC: %v", err)
			}
			responseChan <- &response

		}(contact)
	}

	for _, contact := range candidates.GetContacts(alpha) {
		goSendFindValueRPC(contact)
	}

	// Process responses, make stuff here in general functions, very similar code
	for {
		select {
		case response := <-responseChan:
			activeRPCs--

			// If a value is found, return it and stop the search
			if response.Value != "" {
				value = response.Value
				doneChan <- struct{}{}
				return value, nil
			}

			// If the node returned closer contacts, add them to the candidates list
			for _, contact := range response.ClosestContacts {
				if !visitedNodes[*contact.ID] {
					candidates.Append([]Contact{contact})
					visitedNodes[*contact.ID] = true
					if len(candidates.contacts) < IDLength {
						goSendFindValueRPC(contact)
					}
				}
			}

		default:
			if activeRPCs == 0 {
				doneChan <- struct{}{}
				candidates.Sort()
				return "", candidates.GetContacts(IDLength) // Search did NOT result in a found value, return closest contacts
			}
		}
	}
}

func (kademlia *Kademlia) SendFindValueRPC(contact *Contact, valueID *KademliaID) (FindValueResponse, error) {
	serializedValueID, err := SerializeKademliaID(valueID)
	if err != nil {
		log.Printf("Failed to serialize the KademliaID of value: %v", err)
		return FindValueResponse{}, err
	}

	data, err := kademlia.Network.SendMessageAndWait(contact, FIND_VALUE, REQUEST, serializedValueID)
	if err != nil {
		log.Printf("Failed to send FIND_VALUE RPC: %v", err)
		return FindValueResponse{}, err
	}

	response := FindValueResponse{}

	// Try parse response for value then contacts
	response.Value = string(data.Data)
	if len(response.Value) > 0 {
		return response, nil
	}
	contacts, err := DeserializeContacts(data.Data)
	if err != nil {
		log.Printf("Failed to deserialize contacts: %v", err)
		return FindValueResponse{}, err
	}
	response.ClosestContacts = contacts
	return response, nil
}

func SerializeKademliaID(id *KademliaID) ([]byte, error) {
	if id == nil {
		return nil, fmt.Errorf("cannot serialize nil KademliaID")
	}
	return id[:], nil
}

func DeserializeKademliaID(data []byte) (*KademliaID, error) {
	if len(data) != IDLength {
		return nil, fmt.Errorf("invalid KademliaID length: expected %d, got %d", IDLength, len(data))
	}
	var id KademliaID
	copy(id[:], data)
	return &id, nil
}

func (kademlia *Kademlia) ProcessFindValueMessage(data *[]byte) ([]byte, error) {
	valueID, err := DeserializeKademliaID(*data)
	if err != nil {
		log.Printf("Failed to deserialize value ID: %v", err)
		return nil, err
	}

	// Check if the value is stored locally.
	value, exists := kademlia.Storage[*valueID]
	if exists {
		serializedValue := []byte(value)
		return serializedValue, nil
	}

	// If the value is not found, return the closest contacts to the value ID.
	closestContacts := kademlia.Routes.FindClosestContacts(valueID, IDLength)
	responseData, err := SerializeContacts(closestContacts)
	if err != nil {
		log.Printf("Failed to serialize closest contacts: %v", err)
		return nil, err
	}

	return responseData, nil
}

func (kademlia *Kademlia) Store(data []byte) {
	hexEncodedKey := kademlia.HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)
	//fmt.Printf("Data hash (key): %s\n", hexEncodedKey)

	targetContact := NewContact(kademliaID, "")

	closestContacts := kademlia.LookupContact(&targetContact)

	for _, contact := range closestContacts {
		go func(contact Contact) {
			err := kademlia.SendStoreRPC(&contact, kademliaID, string(data))
			if err != nil {
				log.Printf("Failed to send STORE RPC to %s: %v", contact.Address, err)
			}
		}(contact)
	}
}

func (kademlia *Kademlia) HashData(data string) (hexEncodedKey string) {
	key := sha1.New()
	key.Write([]byte(data))
	hexEncodedKey = hex.EncodeToString(key.Sum(nil))
	return
}

func SerializeSingleContact(contact Contact) ([]byte, error) {
	buffer := new(bytes.Buffer)

	// Serialize the KademliaID (assuming KademliaID is a struct or type that implements binary encoding)
	if err := binary.Write(buffer, binary.LittleEndian, contact.ID); err != nil {
		return nil, fmt.Errorf("failed to serialize KademliaID: %v", err)
	}

	// Serialize the Address length as uint8
	addressLength := uint8(len(contact.Address))
	if err := binary.Write(buffer, binary.LittleEndian, addressLength); err != nil {
		return nil, fmt.Errorf("failed to serialize address length: %v", err)
	}

	// Serialize the Address itself (as bytes)
	if _, err := buffer.Write([]byte(contact.Address)); err != nil {
		return nil, fmt.Errorf("failed to serialize address: %v", err)
	}

	return buffer.Bytes(), nil
}

func SerializeContacts(contacts []Contact) ([]byte, error) {
	buffer := new(bytes.Buffer)

	for _, contact := range contacts {
		contactBytes, err := SerializeSingleContact(contact)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize contact: %v", err)
		}
		buffer.Write(contactBytes)
	}

	return buffer.Bytes(), nil
}

func DeserializeContacts(b []byte) ([]Contact, error) {
	var contacts []Contact
	data := b

	for len(data) > 0 {
		contact, err := DeserializeSingleContact(data)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize contact: %v", err)
		}
		contacts = append(contacts, contact)

		contactSize := binary.Size(*contact.ID) + 1 + len(contact.Address) // ID size + 1 byte for address length + address size
		if len(data) < contactSize {
			return nil, fmt.Errorf("insufficient data for the next contact")
		}
		data = data[contactSize:] // Move forward by the size of the contact just deserialized
	}

	return contacts, nil
}

func DeserializeSingleContact(b []byte) (Contact, error) {
	var contact Contact
	var id KademliaID
	buffer := bytes.NewBuffer(b)

	if err := binary.Read(buffer, binary.LittleEndian, &id); err != nil {
		return contact, fmt.Errorf("failed to deserialize KademliaID: %v", err)
	}
	contact.ID = &id

	var addressLength uint8
	if err := binary.Read(buffer, binary.LittleEndian, &addressLength); err != nil {
		return contact, fmt.Errorf("failed to deserialize address length: %v", err)
	}

	addressBytes := make([]byte, addressLength)
	if _, err := buffer.Read(addressBytes); err != nil {
		return contact, fmt.Errorf("failed to deserialize address: %v", err)
	}
	contact.Address = string(addressBytes)

	return contact, nil
}
