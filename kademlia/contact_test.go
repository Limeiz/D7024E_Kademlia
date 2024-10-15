package kademlia

import (
	"fmt"
	"testing"
)

func TestContact_CalcDistance(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")

	contact1 := NewContact(id1, "127.0.0.1:8080")

	contact1.CalcDistance(id2)

	if contact1.distance == nil {
		t.Errorf("Expected contact1 distance to be calculated, got nil")
	}
}
func TestContact_Less(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")
	id3 := NewKademliaID("1111111111111111111111111111111111111110")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	contact1.CalcDistance(id3)
	contact2.CalcDistance(id3)

	if !contact2.Less(&contact1) {
		t.Errorf("Expected contact2 to be less than contact1")
	}
}

func TestNewContact(t *testing.T) {
	id := NewKademliaID("0000000000000000000000000000000000000000")
	address := "127.0.0.1:8080"
	contact := NewContact(id, address)

	if contact.ID.String() != "0000000000000000000000000000000000000000" {
		t.Errorf("Expected ID to be '0000000000000000000000000000000000000000', got %s", contact.ID)
	}
	if contact.Address != "127.0.0.1:8080" {
		t.Errorf("Expected address to be '127.0.0.1:8080', got %s", contact.Address)
	}
}

func TestContact_String(t *testing.T) {
	id := NewKademliaID("0000000000000000000000000000000000000000")
	contact := NewContact(id, "127.0.0.1:8080")

	expectedString := fmt.Sprintf(`contact("%s", "%s")`, id.String(), "127.0.0.1:8080")
	if contact.String() != expectedString {
		t.Errorf("Expected string representation to be %s, got %s", expectedString, contact.String())
	}
}

// Test Append and GetContacts in ContactCandidates
func TestContactCandidates_Append(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	candidates.Append([]Contact{contact1, contact2})

	if len(candidates.contacts) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(candidates.contacts))
	}
}

func TestContactCandidates_GetContacts(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	candidates.Append([]Contact{contact1, contact2})

	result := candidates.GetContacts(1)
	if len(result) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(result))
	}

	if result[0] != contact1 {
		t.Errorf("Expected first contact to be contact1, got %v", result[0])
	}
}

func TestContactCandidates_Sort(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")
	id3 := NewKademliaID("1111111111111111111111111111111111111110")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	candidates.Append([]Contact{contact2, contact1})

	contact1.CalcDistance(id3)
	contact2.CalcDistance(id3)

	// candidates.Sort()
	// sortedContacts := candidates.GetContacts(2)

	// if !sortedContacts[0].Less(&sortedContacts[1]) {
	// 	t.Errorf("Expected contacts to be sorted, but they are not")
	// }
	// if sortedContacts[0].distance == nil || sortedContacts[1].distance == nil {
	// 	t.Error("Distance for one or more contacts is nil")
	// }

}

func TestContactCandidates_Len(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	candidates.Append([]Contact{contact1, contact2})

	if candidates.Len() != 2 {
		t.Errorf("Expected length to be 2, got %d", candidates.Len())
	}
}

func TestContactCandidates_Swap(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	candidates.Append([]Contact{contact1, contact2})

	// Swap the two contacts
	candidates.Swap(0, 1)

	if candidates.contacts[0].Address != contact2.Address {
		t.Errorf("Expected first contact to be contact2, got %v", candidates.contacts[0])
	}

	if candidates.contacts[1].Address != contact1.Address {
		t.Errorf("Expected second contact to be contact1, got %v", candidates.contacts[1])
	}
}

func TestContactCandidates_Less(t *testing.T) {
	candidates := &ContactCandidates{}
	id1 := NewKademliaID("0000000000000000000000000000000000000000")
	id2 := NewKademliaID("1111111111111111111111111111111111111111")
	id3 := NewKademliaID("1111111111111111111111111111111111111110")

	contact1 := NewContact(id1, "127.0.0.1:8080")
	contact2 := NewContact(id2, "127.0.0.1:8081")

	contact1.CalcDistance(id3)
	contact2.CalcDistance(id3)

	candidates.Append([]Contact{contact2, contact1})

	if !candidates.Less(0, 1) {
		t.Errorf("Expected contact2 to be less than contact1")
	}

	candidates.Swap(0, 1)

	if candidates.Less(0, 1) {
		t.Errorf("Expected contact1 to be less than contact2 after swap")
	}
}
