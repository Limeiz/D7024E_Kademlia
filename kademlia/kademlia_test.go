package kademlia

import (
	"testing"
)

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
