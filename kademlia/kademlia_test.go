package kademlia

import (
	"os"
	"testing"
	"time"
)

func NewMockKademlia() *Kademlia {
	me := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	contact := NewContact(me, "127.0.0.1:8080")
	return &Kademlia{
		Storage: make(map[KademliaID]*StorageItem),
		Routes:  NewRoutingTable(contact),
	}
}

func TestInitNode(t *testing.T) {
	os.Setenv("OBJECT_TTL", "10")
	os.Setenv("NODE_TYPE", "bootstrap")
	os.Setenv("BOOTSTRAP_NODE", "127.0.0.1:8080")

	bootstrapID := NewKademliaID("2111111400000000000000000000000000000000")

	kademlia := InitNode(bootstrapID)

	if kademlia == nil {
		t.Errorf("Expected kademlia node to be initialized, but got nil.")
		return
	}

	expectedTTL := 10 * time.Second
	if kademlia.TTL != expectedTTL {
		t.Errorf("Expected TTL to be %v, got %v", expectedTTL, kademlia.TTL)
	}

	if len(kademlia.Storage) != 0 {
		t.Errorf("Expected Storage to be initialized as an empty map, got %v", kademlia.Storage)
	}

	os.Unsetenv("OBJECT_TTL")
	os.Unsetenv("NODE_TYPE")
	os.Unsetenv("BOOTSTRAP_NODE")
}

func TestInitNodeInvalidTTL(t *testing.T) {
	os.Setenv("OBJECT_TTL", "invalid_ttl")

	bootstrapID := NewKademliaID("2111111400000000000000000000000000000000")
	kademlia := InitNode(bootstrapID)

	if kademlia != nil {
		t.Errorf("Expected kademlia node to be nil due to invalid TTL, but got: %+v", kademlia)
	}

	os.Unsetenv("OBJECT_TTL")
}

// func TestHashSerializeAndDeserializeData(t *testing.T) {
// 	data := "hej"
// 	hexEncodedKey := HashData(string(data))
// 	kademliaID := NewKademliaID(hexEncodedKey)

// 	// Serialize the KademliaID
// 	serializedData, err := SerializeData(kademliaID)
// 	if err != nil {
// 		t.Fatalf("Failed to serialize data: %v", err)
// 	}

// 	// Deserialize
// 	var deserializedID KademliaID
// 	deserializedID, err = DeserializeData[KademliaID](serializedData)
// 	if err != nil {
// 		t.Fatalf("Failed to deserialize data: %v", err)
// 	}

// 	// Check if the original ID and the deserialized ID are equal
// 	if *kademliaID != deserializedID {
// 		t.Errorf("Expected %v, but got %v", kademliaID, deserializedID)
// 	}
// }

func TestHashSerializeAndDeserializeData(t *testing.T) {
	data := "hej"
	hexEncodedKey := HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)

	// Successful Serialization Test
	serializedData, err := SerializeData(kademliaID)
	if err != nil {
		t.Fatalf("Failed to serialize data: %v", err)
	}

	// Successful Deserialization Test
	var deserializedID KademliaID
	deserializedID, err = DeserializeData[KademliaID](serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize data: %v", err)
	}

	// Check if the original ID and the deserialized ID are equal
	if *kademliaID != deserializedID {
		t.Errorf("Expected %v, but got %v", kademliaID, deserializedID)
	}

	// Test Error Handling for DeserializeData with corrupted data
	corruptedData := []byte{0x00, 0x01, 0x02, 0x03, 0x04} // Random invalid bytes
	_, err = DeserializeData[KademliaID](corruptedData)
	if err == nil {
		t.Fatal("Expected error during deserialization of corrupted data, but got none")
	}

	// Test Error Handling for SerializeData with a nil value
	var nilKademliaID *KademliaID
	_, err = SerializeData(nilKademliaID)
	if err == nil {
		t.Fatal("Expected error during serialization of nil data, but got none")
	}

	// Test Error Handling for DeserializeData with empty data
	emptyData := []byte{}
	_, err = DeserializeData[KademliaID](emptyData)
	if err == nil {
		t.Fatal("Expected error during deserialization of empty data, but got none")
	}

	type UnsupportedType struct {
		ch chan int // gob cannot encode channels
	}

	unsupportedData := UnsupportedType{ch: make(chan int)}
	_, err = SerializeData(unsupportedData)
	if err == nil {
		t.Fatal("Expected error during serialization of unsupported type, but got none")
	} else {
		t.Logf("Correctly failed to serialize unsupported data: %v", err)
	}
}

func TestStorage(t *testing.T) {
	bootstrapID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	data := "aa"
	os.Setenv("OBJECT_TTL", "10")

	kademlia := InitNode(bootstrapID)
	hexEncodedKey := HashData(string(data))
	kademliaID := NewKademliaID(hexEncodedKey)
	//fmt.Printf("Data hash (key): %s\n", hexEncodedKey)

	kademlia.StorageSet(kademliaID, &data)

	storedValue, exists := kademlia.StorageGet(kademliaID)
	if !exists {
		t.Error("Expected data to exist in storage, but it does not")
	}

	if storedValue != data {
		t.Errorf("Expected stored value to be '%s', got '%s'", data, storedValue)
	}

}

func TestRecieveStoreRPC(t *testing.T) {
	bootstrapID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	os.Setenv("OBJECT_TTL", "10")
	kademlia := InitNode(bootstrapID)

	value := "mockData"
	hexEncodedKey := HashData(string(value))
	key := NewKademliaID(hexEncodedKey)
	storeData := StoreData{Key: *key, Value: value}

	serializedData, err := SerializeData(storeData)
	if err != nil {
		t.Fatalf("Failed to serialize data: %v", err)
	}

	err = kademlia.RecieveStoreRPC(&serializedData)
	if err != nil {
		t.Fatalf("RecieveStoreRPC returned an error: %v", err)
	}

	storedValue, exists := kademlia.StorageGet(key)

	if !exists {
		t.Errorf("Expected key %s to be stored, but it was not found.", key.String())
	}
	if storedValue != value {
		t.Errorf("Expected stored value to be %s, got %s.", value, storedValue)
	}
}

func TestSerializeAndDeserializeContact(t *testing.T) {
	contactID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	contact := NewContact(contactID, "127.0.0.1:8080")

	serializedData, err := SerializeSingleContact(contact)
	if err != nil {
		t.Fatalf("Failed to serialize contact: %v", err)
	}

	deserializedContact, err := DeserializeSingleContact(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize contact: %v", err)
	}

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
	kademlia.StorageSet(key, &value)

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

func TestSerializeKademliaID(t *testing.T) {
	tests := []struct {
		id       *KademliaID
		expected []byte
		hasError bool
	}{
		{ // Test case with a valid KademliaID
			id:       &KademliaID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expected: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			hasError: false,
		},
		{ // Test case with a nil KademliaID
			id:       nil,
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := SerializeKademliaID(test.id)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for SerializeKademliaID with id %v, got nil", test.id)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for SerializeKademliaID with id %v: %v", test.id, err)
			}
			if !equalbyte(result, test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		}
	}
}

func TestDeserializeKademliaID(t *testing.T) {
	tests := []struct {
		data     []byte
		expected *KademliaID
		hasError bool
	}{
		{ // Test case with valid data
			data:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expected: &KademliaID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			hasError: false,
		},
		{ // Test case with invalid length
			data:     []byte{1, 2, 3}, // Invalid length
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := DeserializeKademliaID(test.data)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for DeserializeKademliaID with data %v, got nil", test.data)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for DeserializeKademliaID with data %v: %v", test.data, err)
			}
			if !equalID(result, test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		}
	}
}

// Helper function to compare two KademliaID values
func equalID(a *KademliaID, b *KademliaID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	for i := 0; i < IDLength; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalbyte(a []byte, b []byte) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	for i := 0; i < IDLength; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestProcessFindValueMessage(t *testing.T) {
	kademlia := NewMockKademlia()
	value := "storedValue"
	hash := HashData(string(value))
	key := NewKademliaID(hash)
	kademlia.StorageSet(key, &value)

	data := make([]byte, 20)
	copy(data, key[:])

	// Valid
	result, err := kademlia.ProcessFindValueMessage(&data)
	if err != nil {
		t.Fatalf("Unexpected error for valid data: %v", err)
	}
	if string(result) != "storedValue" {
		t.Errorf("Expected value %q, got %q", "storedValue", string(result))
	}

	// Invalid KademliaID length
	invalidData := []byte{1, 2, 3, 4, 5}
	_, err = kademlia.ProcessFindValueMessage(&invalidData)
	if err == nil {
		t.Fatalf("Expected error for data %v, got none", invalidData)
	}
	if err.Error() != "invalid KademliaID length: expected 20, got 5\n" {
		t.Errorf("Expected error message for invalid length, got %v", err)
	}

	// When value not found
	missingData := make([]byte, 20)
	missingID := NewKademliaID(HashData(string(missingData)))

	copy(missingData, missingID[:])

	contact1 := NewContact(NewRandomKademliaID(), "localhost:8002")
	kademlia.Routes.AddContact(contact1)

	responseData, err := kademlia.ProcessFindValueMessage(&missingData)
	if err != nil {
		t.Fatalf("Unexpected error for missing data: %v", err)
	}

	if responseData == nil {
		t.Fatal("Expected response data to be non-nil")
	}

	expectedContacts := []Contact{contact1}
	expectedResponse, err := SerializeContacts(expectedContacts)
	if err != nil {
		t.Fatalf("Failed to serialize expected contacts: %v", err)
	}

	if string(responseData) != string(expectedResponse) {
		t.Errorf("Expected response data to be %q, got %q", string(expectedResponse), string(responseData))
	}
}

func TestSendStoreRPC_Serialization(t *testing.T) {
	kademlia := NewMockKademlia()
	contact := &Contact{
		ID:      NewRandomKademliaID(),
		Address: "127.0.0.1:8000",
	}

	kademlia.Network = createMockNetwork().Node.Network
	data := "Some test data"
	key := NewKademliaID(HashData(data))

	err := kademlia.SendStoreRPC(contact, key, data)
	if err == nil || err.Error() != "Error: Store could not be sent to 127.0.0.1:8000 \n" {
		t.Errorf("Expected Store could not be sent error, got: %v", err)
	}
}

func TestPing_Failure(t *testing.T) {
	kademlia := NewMockKademlia()
	contact := &Contact{
		ID:      NewRandomKademliaID(),
		Address: "127.0.0.1:8000",
	}

	kademlia.Network = createMockNetwork().Node.Network

	kademlia.Routes.AddContact(*contact)
	if !kademlia.Routes.IsContactInTable(contact) {
		t.Errorf("Expected contact %v to be in routing table after adding", contact.ID)
	}

	// Call the Ping method
	err := kademlia.Ping(contact)
	if err == nil || err.Error() != "Error: Ping could not be sent to 127.0.0.1:8000\n" {
		t.Errorf("Expected Ping error, got: %v", err)
	}

	// // Check if the contact was removed from the routing table
	// if kademlia.Routes.IsContactInTable(contact) {
	// 	t.Errorf("Expected contact %v to be removed from routing table after ping failure", contact.ID)
	// }
}

func TestStorageExists(t *testing.T) {
	bootstrapID := NewKademliaID("FABFABFABFABFABFABFABFABFABFABFABFABFAB0")
	os.Setenv("OBJECT_TTL", "10")
	kademlia := InitNode(bootstrapID)

	value := "mockData"
	hexEncodedKey := HashData(string(value))
	key := NewKademliaID(hexEncodedKey)

	// Test case 1: Key does not exist
	exists := kademlia.StorageExists(key)
	if exists {
		t.Errorf("Expected key %s to not exist, but found it in storage", key.String())
	}

	// Store the key in Kademlia's storage
	kademlia.StorageSet(key, &value)

	// Test case 2: Key exists
	exists = kademlia.StorageExists(key)
	if !exists {
		t.Errorf("Expected key %s to exist in storage, but it was not found", key.String())
	}
}

func TestStorageItemUsage(t *testing.T) {
	item := StorageItem{
		Value:      "example value",
		Expiration: time.Now().Add(10 * time.Second), // Set expiration to 10 seconds from now
	}

	if item.Value != "example value" {
		t.Errorf("Expected item.Value to be 'example value', got '%s'", item.Value)
	}

	if item.Expiration.Before(time.Now()) {
		t.Error("Expected item.Expiration to be in the future, but it is not")
	}
}
