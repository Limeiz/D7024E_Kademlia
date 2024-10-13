package kademlia

import (
	"testing"
)

func NewMockKademlia() *Kademlia {
	me := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	contact := NewContact(me, "127.0.0.1:8080")
	return &Kademlia{
		Storage: make(map[KademliaID]string),
		Routes:  NewRoutingTable(contact),
	}
}

func TestInitNode(t *testing.T) {
	bootstrapID := NewKademliaID("2111111400000000000000000000000000000000")
	kademlia := InitNode(bootstrapID)

	if kademlia == nil {
		t.Errorf("Expected kademlia node to be initialized, but got nil.")
	}
}

func TestHashSerializeAndDeserializeData(t *testing.T) {
	data := "hej"
	hexEncodedKey := HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)

	// Serialize the KademliaID
	serializedData, err := SerializeData(kademliaID)
	if err != nil {
		t.Fatalf("Failed to serialize data: %v", err)
	}

	// Deserialize
	var deserializedID KademliaID
	deserializedID, err = DeserializeData[KademliaID](serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize data: %v", err)
	}

	// Check if the original ID and the deserialized ID are equal
	if *kademliaID != deserializedID {
		t.Errorf("Expected %v, but got %v", kademliaID, deserializedID)
	}
}

func TestStorage(t *testing.T) {
	bootstrapID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	data := "aa"
	kademlia := InitNode(bootstrapID)
	hexEncodedKey := HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)
	//fmt.Printf("Data hash (key): %s\n", hexEncodedKey)

	kademlia.Storage[*kademliaID] = string(data)

	storedValue, exists := kademlia.Storage[*kademliaID]
	if !exists {
		t.Error("Expected data to exist in storage, but it does not")
	}

	// Verify that the stored value matches the original data
	if storedValue != data {
		t.Errorf("Expected stored value to be '%s', got '%s'", data, storedValue)
	}

}

func TestRecieveStoreRPC(t *testing.T) {
	bootstrapID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	kademlia := InitNode(bootstrapID)

	value := "mockData"
	hexEncodedKey := HashData(string(value))
	key := NewKademliaID(hexEncodedKey)
	storeData := StoreData{Key: *key, Value: value}

	// Serialize the mock data
	serializedData, err := SerializeData(storeData)
	if err != nil {
		t.Fatalf("Failed to serialize data: %v", err)
	}

	// Call the RecieveStoreRPC method
	err = kademlia.RecieveStoreRPC(&serializedData)
	if err != nil {
		t.Fatalf("RecieveStoreRPC returned an error: %v", err)
	}

	// Check if the data was stored correctly
	kademlia.StorageMapMutex.Lock()
	storedValue, exists := kademlia.Storage[*key]
	kademlia.StorageMapMutex.Unlock()

	if !exists {
		t.Errorf("Expected key %s to be stored, but it was not found.", key.String())
	}
	if storedValue != value {
		t.Errorf("Expected stored value to be %s, got %s.", value, storedValue)
	}
}

func TestSerializeAndDeserializeContact(t *testing.T) {
	// Create a contact
	contactID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	contact := NewContact(contactID, "127.0.0.1:8080")

	// Serialize the contact
	serializedData, err := SerializeSingleContact(contact)
	if err != nil {
		t.Fatalf("Failed to serialize contact: %v", err)
	}

	// Deserialize the contact
	deserializedContact, err := DeserializeSingleContact(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize contact: %v", err)
	}

	// Check if the original contact and the deserialized contact are equal
	if *contact.ID != *deserializedContact.ID || contact.Address != deserializedContact.Address {
		t.Errorf("Expected deserialized contact to be %+v, but got %+v", contact, deserializedContact)
	}
}

func TestProcessFindContactMessage(t *testing.T) {
	me := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	contact := NewContact(me, "127.0.0.1:8080")
	kademlia := &Kademlia{
		Routes: NewRoutingTable(contact),
	}

	targetID := NewRandomKademliaID()
	targetContact := Contact{ID: targetID, Address: "localhost:8001"}
	kademlia.Routes.AddContact(targetContact)

	serializedTarget, err := SerializeSingleContact(targetContact)
	if err != nil {
		t.Fatalf("Failed to serialize target contact: %v", err)
	}

	senderID := NewRandomKademliaID()
	senderContact := Contact{ID: senderID, Address: "localhost:8002"}

	response, err := kademlia.ProcessFindContactMessage(&serializedTarget, senderContact)
	if err != nil {
		t.Fatalf("ProcessFindContactMessage returned an error: %v", err)
	}

	if !kademlia.Routes.IsContactInTable(&senderContact) {
		t.Errorf("Expected sender contact %s to be added to routing table", senderContact.ID.String())
	}

	closestContacts, err := DeserializeContacts(response)
	if err != nil {
		t.Fatalf("Failed to deserialize response: %v", err)
	}

	if len(closestContacts) == 0 {
		t.Errorf("Expected closest contacts, but got none")
	}
}

func TestLookupData_ExistingData(t *testing.T) {
	kademlia := NewMockKademlia()

	value := "mockData"
	hash := HashData(string(value))
	key := NewKademliaID(hash)
	kademlia.Storage[*key] = string(value)

	data, contacts, err := kademlia.LookupData(hash)
	if err != nil {
		t.Fatalf("LookupData returned an error: %v", err)
	}

	if data != "mockData" {
		t.Errorf("Expected data to be 'mockData', got '%s'", data)
	}

	if len(contacts) == 0 || contacts[0].ID.String() != kademlia.Routes.Me.ID.String() {
		t.Errorf("Expected contact to be present, got: %+v", contacts)
	}
}

func TestLookupData_NonExistentData(t *testing.T) {
	kademlia := NewMockKademlia()
	hash := "nonexistenthash"

	data, contacts, err := kademlia.LookupData(hash)
	if err == nil {
		t.Fatal("Expected error for nonexistent hash, but got none")
	}

	if data != "" {
		t.Errorf("Expected data to be empty, got '%s'", data)
	}

	if len(contacts) != 0 {
		t.Errorf("Expected no contacts, got: %+v", contacts)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},       // a < b
		{2, 1, 1},       // a > b
		{3, 3, 3},       // a == b
		{0, 0, 0},       // both zero
		{-1, 1, -1},     // negative vs positive
		{5, -5, -5},     // positive vs negative
		{-10, -20, -20}, // two negatives
	}

	for _, test := range tests {
		result := min(test.a, test.b)
		if result != test.expected {
			t.Errorf("min(%d, %d) = %d; expected %d", test.a, test.b, result, test.expected)
		}
	}
}
