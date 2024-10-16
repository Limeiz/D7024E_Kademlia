package kademlia

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

const alpha = 3

type Kademlia struct {
	Routes          *RoutingTable
	Network         *Network
	Storage         map[KademliaID]string
	StorageMapMutex sync.Mutex
	ShutdownChan    chan struct{}
}

type FindValueResponse struct {
	Value           string
	ClosestContacts []Contact // The list of closer contacts, if no value is found
}

type StoreData struct {
	Key   KademliaID
	Value string
}

// Please use these instead of creating a seriliaze and deserialize for every type of struct
func SerializeData[T any](data T) ([]byte, error) {
	if isNil(data) {
		return nil, fmt.Errorf("cannot serialize nil data")
	}
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func isNil[T any](v T) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()
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
		Storage:      make(map[KademliaID]string),
		ShutdownChan: make(chan struct{}),
	}
	bootstrap_contact := NewContact(bootstrap_id, os.Getenv("BOOTSTRAP_NODE"))
	if os.Getenv("NODE_TYPE") == "bootstrap" {
		kademlia_node.Routes = NewRoutingTable(bootstrap_contact)
	} else {
		new_id := NewRandomKademliaID()
		new_me_contact := NewContact(new_id, GetLocalIP())
		kademlia_node.Routes = NewRoutingTable(new_me_contact)
		kademlia_node.Routes.AddContact(bootstrap_contact)
	}

	return &kademlia_node
}

func (kademlia *Kademlia) InitNetwork(target *Contact) {
	attempt := 0
	// Sometimes when it doesn't work, kick it in the butt again
	// UDP can miss some messages since it get congested, so refer to message above
	for {
		log.Printf("LookupContact attempt %d \n", attempt+1)
		contacts, error := kademlia.LookupContact(target)
		log.Printf("LookupContact attempt %d finished \n", attempt+1)
		log.Printf("LookupContact produced %d contacts \n", len(contacts))
		if error != nil || len(contacts) < 2 {
			log.Printf("Error: LookupContact attempt %d failed \n", attempt+1)
			delay := 5.00 + rand.Float64()*(10.00-5.01) // wait for 5-10 seconds
			time.Sleep(time.Duration(delay*1000) * time.Millisecond)
			attempt++
			continue
		}
		break
	}
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
		return fmt.Errorf("Error: Could not serialize store data! \n")
	}
	_, err := kademlia.Network.SendMessageAndWait(contact, STORE, REQUEST, serialized_data)
	if err != nil {
		log.Printf("Error: Store could not be sent to %s\n", contact.Address)
		return fmt.Errorf("Error: Store could not be sent to %s \n", contact.Address)
	}

	return nil

}

func (kademlia *Kademlia) RecieveStoreRPC(data *[]byte) error {
	deserialized_data, err := DeserializeData[StoreData](*data)
	if err != nil {
		log.Printf("Error: Could not deserialize store data")
		return err
	}
	kademlia.StorageMapMutex.Lock()
	kademlia.Storage[deserialized_data.Key] = deserialized_data.Value
	kademlia.StorageMapMutex.Unlock()
	log.Printf("Data stored in node %s\n", kademlia.Routes.Me.ID.String())
	return nil
}

func (kademlia *Kademlia) LookupContact(target *Contact) ([]Contact, error) {
	done := make(chan []Contact, bucketSize)
	ret := make([]Contact, 0, bucketSize)
	frontier := make([]Contact, 0) // shortlist
	seen := make(map[KademliaID]bool)

	initialContacts := kademlia.Routes.FindClosestContacts(target.ID, bucketSize)
	if len(initialContacts) == 0 {
		log.Println("No contacts found in the routing table.")
		return nil, fmt.Errorf("No contacts found in the routing table.")
	}
	for _, contact := range initialContacts {
		ret = append(ret, contact)
		frontier = append(frontier, contact)
		seen[*contact.ID] = true
	}

	pending := 0
	for i := 0; i < alpha && len(frontier) > 0; i++ {
		sort.Slice(frontier, func(i, j int) bool {
			return frontier[i].Less(&frontier[j])
		})
		pending++
		contact := frontier[0]
		frontier = frontier[1:]
		go kademlia.SendFindNode(contact, target, done)
	}

	for pending > 0 {
		nodes := <-done
		pending--
		if nodes == nil {
			continue
		}
		for _, node := range nodes {
			if !seen[*node.ID] {
				seen[*node.ID] = true
				node.CalcDistance(target.ID)
				ret = append(ret, node)
				frontier = append(frontier, node)
			}
		}

		for len(frontier) > 0 { //pending < alpha &&
			sort.Slice(frontier, func(i, j int) bool {
				return frontier[i].Less(&frontier[j])
			})
			pending++
			contact := frontier[0]
			frontier = frontier[1:]
			go kademlia.SendFindNode(contact, target, done)
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Less(&ret[j])
	})

	if len(ret) > bucketSize {
		ret = ret[:bucketSize]
	}

	return ret, nil
}

func (kademlia *Kademlia) SendFindNode(contact Contact, target *Contact, done chan<- []Contact) {
	response, err := kademlia.SendFindNodeRPC(&contact, target)
	if err != nil {
		log.Printf("Error contacting %v: %v ", contact.Address, err)
		if !strings.EqualFold(contact.ID.String(), os.Getenv("BOOTSTRAP_ID")) { // dont send ping to bootstrap
			kademlia.Ping(&contact) // Ping will remove the contact if it doesnt work
		}
		done <- nil
		return
	}
	done <- response
}

func (kademlia *Kademlia) SendFindNodeRPC(contact *Contact, target *Contact) ([]Contact, error) {
	serializedTarget, err := SerializeSingleContact(*target)
	if err != nil {
		log.Printf("Failed to serialize target contact: %v", err)
		return nil, err
	}

	data, err := kademlia.Network.SendMessageAndWait(contact, FIND_NODE, REQUEST, serializedTarget)
	if err != nil {
		log.Printf("Failed to send FIND_NODE RPC: %v \n", err)
		return nil, err
	}

	contacts, err := DeserializeContacts(data.Data)
	if err != nil {
		log.Printf("Failed to deserialize contacts: %v \n", err)
		return nil, err
	}

	log.Printf("Received response from contact: %s with %d contacts \n", contact.ID.String(), len(contacts))
	return contacts, nil
}

func (kademlia *Kademlia) ProcessFindContactMessage(data *[]byte, sender Contact) ([]byte, error) {
	target, err := DeserializeSingleContact(*data) // assumes msg data only holds target contact
	if err != nil {
		log.Printf("Failed to deserialize target contact: %v \n", err)
		return nil, err
	}
	targetID := target.ID

	closestContacts := kademlia.Routes.FindClosestContacts(targetID, IDLength)

	kademlia.Routes.AddContact(sender)

	// Serialize the list of closest contacts
	responseBytes, err := SerializeContacts(closestContacts)
	if err != nil {
		log.Printf("Failed to serialize contacts: %v \n", err)
		return nil, err
	}

	log.Printf("Sent closest contacts from ProcessFindContactMessage \n")
	return responseBytes, err
}

func (kademlia *Kademlia) LookupData(hash string) (string, []Contact, error) {
	kademliaID := NewKademliaID(hash)
	if kademliaID == nil {
		return "", nil, fmt.Errorf("Invalid KademliaID for hash: %s", hash)
	}
	kademlia.StorageMapMutex.Lock()
	data, exists := kademlia.Storage[*kademliaID]
	kademlia.StorageMapMutex.Unlock()

	if exists {
		return data, []Contact{kademlia.Routes.Me}, nil
	}

	type FindValueResult struct {
		response FindValueResponse
		err      error
	}

	done := make(chan FindValueResult)
	ret := make([]Contact, 0, bucketSize)
	frontier := make([]Contact, 0)
	seen := make(map[KademliaID]bool)
	targetID := NewKademliaID(hash)

	initialContacts := kademlia.Routes.FindClosestContacts(targetID, bucketSize)
	if len(initialContacts) == 0 {
		log.Println("No contacts found in the routing table.")
		return "", nil, fmt.Errorf("No contacts found in the routing table.")
	}
	for _, contact := range initialContacts {
		ret = append(ret, contact)
		frontier = append(frontier, contact)
		seen[*contact.ID] = true
	}

	pending := 0
	var mu sync.Mutex

	sendFindValue := func(contact Contact) {
		go func(c Contact) {
			response, err := kademlia.SendFindValueRPC(&c, targetID)
			done <- FindValueResult{response: response, err: err}
		}(contact)
	}

	mu.Lock()
	alphaContacts := min(alpha, len(frontier))
	for i := 0; i < alphaContacts; i++ {
		contact := frontier[0]
		frontier = frontier[1:]
		pending++
		sendFindValue(contact)
	}
	mu.Unlock()

	for {
		mu.Lock()
		if pending == 0 && len(frontier) == 0 {
			mu.Unlock()
			break
		}
		mu.Unlock()

		select {
		case res := <-done:
			mu.Lock()
			pending--
			mu.Unlock()

			if res.err != nil {
				log.Printf("Error contacting node: %v", res.err)
				continue
			}

			if res.response.Value != "" {
				if len(res.response.ClosestContacts) > 0 { //tHESE THREE
					return res.response.Value, []Contact{res.response.ClosestContacts[0]}, nil
				}
				return res.response.Value, nil, nil
			}
			mu.Lock()
			for _, node := range res.response.ClosestContacts {
				if !seen[*node.ID] {
					seen[*node.ID] = true
					node.CalcDistance(targetID)
					ret = append(ret, node)
					frontier = append(frontier, node)
				}
			}

			for pending < alpha && len(frontier) > 0 {
				contact := frontier[0]
				frontier = frontier[1:]
				pending++
				sendFindValue(contact)
			}
			mu.Unlock()

		case <-time.After(time.Second * 10):
			log.Println("LookupData timed out.")
			return "", ret, fmt.Errorf("LookupData timed out")
		}
	}

	// Finalize the result
	sort.Slice(ret, func(i, j int) bool {
		if i >= len(ret) || j >= len(ret) {
			log.Printf("Index out of range: i=%d, j=%d, len(ret)=%d", i, j, len(ret))
			return false
		}
		return ret[i].Less(&ret[j])
	})

	if len(ret) > bucketSize {
		ret = ret[:bucketSize]
	}

	return "", ret, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (kademlia *Kademlia) SendFindValueRPC(contact *Contact, valueID *KademliaID) (FindValueResponse, error) {
	serializedValueID, err := SerializeKademliaID(valueID)
	if err != nil {
		log.Printf("Failed to serialize the KademliaID of value: %v\n", err)
		return FindValueResponse{}, err
	}

	data, err := kademlia.Network.SendMessageAndWait(contact, FIND_VALUE, REQUEST, serializedValueID)
	if err != nil {
		log.Printf("Failed to send FIND_VALUE RPC: %v\n", err)
		return FindValueResponse{}, err
	}

	response := FindValueResponse{
		ClosestContacts: make([]Contact, 0),
	}

	// Try parse response for value then contacts
	response.Value = string(data.Data)
	if len(response.Value) > 0 {
		response.ClosestContacts = append(response.ClosestContacts, *contact) // also send the node it was found on THIS LINE
		return response, nil
	}
	contacts, err := DeserializeContacts(data.Data)
	if err != nil {
		log.Printf("Failed to deserialize contacts: %v\n", err)
		return FindValueResponse{}, err
	}
	response.ClosestContacts = contacts
	return response, nil
}

func (kademlia *Kademlia) Quit() {
	log.Printf("Quitting... Good bye! \n")
	close(kademlia.ShutdownChan)
}

func SerializeKademliaID(id *KademliaID) ([]byte, error) {
	if id == nil {
		return nil, fmt.Errorf("cannot serialize nil KademliaID\n")
	}
	return id[:], nil
}

func DeserializeKademliaID(data []byte) (*KademliaID, error) {
	if len(data) != IDLength {
		return nil, fmt.Errorf("invalid KademliaID length: expected %d, got %d\n", IDLength, len(data))
	}
	var id KademliaID
	copy(id[:], data)
	return &id, nil
}

func (kademlia *Kademlia) ProcessFindValueMessage(data *[]byte) ([]byte, error) {
	valueID, err := DeserializeKademliaID(*data)
	if err != nil {
		log.Printf("Failed to deserialize value ID: %v\n", err)
		return nil, err
	}

	// Check if the value is stored locally.
	kademlia.StorageMapMutex.Lock()
	value, exists := kademlia.Storage[*valueID]
	kademlia.StorageMapMutex.Unlock()
	if exists {
		serializedValue := []byte(value)
		return serializedValue, nil
	}

	// If the value is not found, return the closest contacts to the value ID.
	closestContacts := kademlia.Routes.FindClosestContacts(valueID, IDLength)
	responseData, err := SerializeContacts(closestContacts)
	if err != nil {
		log.Printf("Failed to serialize closest contacts: %v\n", err)
		return nil, err
	}

	return responseData, nil
}

func (kademlia *Kademlia) Store(data []byte) (string, error) {
	hexEncodedKey := HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)

	targetContact := NewContact(kademliaID, "")

	closestContacts, lookup_err := kademlia.LookupContact(&targetContact)
	if lookup_err != nil {
		return "", lookup_err
	}
	fmt.Printf("Closest contact received: %v\n", closestContacts)

	for _, contact := range closestContacts {
		go func(contact Contact) {
			err := kademlia.SendStoreRPC(&contact, kademliaID, string(data))
			if err != nil {
				log.Printf("Failed to send STORE RPC to %s: %v\n", contact.Address, err)
			}
		}(contact)
	}
	return hexEncodedKey, nil // add error response
}

func SerializeSingleContact(contact Contact) ([]byte, error) {
	buffer := new(bytes.Buffer)

	// Serialize the KademliaID (assuming KademliaID is a struct or type that implements binary encoding)
	if err := binary.Write(buffer, binary.BigEndian, contact.ID); err != nil {
		return nil, fmt.Errorf("failed to serialize KademliaID: %v \n", err)
	}

	// Serialize the Address length as uint8
	addressLength := uint8(len(contact.Address))
	if err := binary.Write(buffer, binary.BigEndian, addressLength); err != nil {
		return nil, fmt.Errorf("failed to serialize address length: %v \n", err)
	}

	// Serialize the Address itself (as bytes)
	if _, err := buffer.Write([]byte(contact.Address)); err != nil {
		return nil, fmt.Errorf("failed to serialize address: %v \n", err)
	}

	return buffer.Bytes(), nil
}

func SerializeContacts(contacts []Contact) ([]byte, error) {
	buffer := new(bytes.Buffer)

	for _, contact := range contacts {
		contactBytes, err := SerializeSingleContact(contact)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize contact: %v \n", err)
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
			return nil, fmt.Errorf("failed to deserialize contact: %v \n", err)
		}
		contacts = append(contacts, contact)

		contactSize := binary.Size(*contact.ID) + 1 + len(contact.Address)
		if len(data) < contactSize {
			return nil, fmt.Errorf("insufficient data for the next contact \n")
		}
		data = data[contactSize:] // Move forward by the size of the contact just deserialized
	}

	return contacts, nil
}

func DeserializeSingleContact(b []byte) (Contact, error) {
	var contact Contact
	var id KademliaID
	buffer := bytes.NewBuffer(b)

	if err := binary.Read(buffer, binary.BigEndian, &id); err != nil {
		return contact, fmt.Errorf("failed to deserialize KademliaID: %v \n", err)
	}
	contact.ID = &id

	var addressLength uint8
	if err := binary.Read(buffer, binary.BigEndian, &addressLength); err != nil {
		return contact, fmt.Errorf("failed to deserialize address length: %v \n", err)
	}

	addressBytes := make([]byte, addressLength)
	if _, err := buffer.Read(addressBytes); err != nil {
		return contact, fmt.Errorf("failed to deserialize address: %v \n", err)
	}
	contact.Address = string(addressBytes)

	return contact, nil
}
