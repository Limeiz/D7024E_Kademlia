package kademlia

import (
	"testing"
)

func TestRoutingTable(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002"))

	contacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	/* for i := range contacts {
		fmt.Println(contacts[i].String())
	} */

	// TODO: This is just an example. Make more meaningful assertions.
	if len(contacts) != 6 {
		t.Fatalf("Expected 6 contacts but instead got %d", len(contacts))
	}
}

func TestNewRoutingTable(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	if rt == nil {
		t.Error("Expected routing table to be non-nil")
	}
	if rt.Me != me {
		t.Errorf("Expected Me to be %v, got %v", me, rt.Me)
	}

	if len(rt.buckets) != IDLength*8 {
		t.Errorf("Expected %d buckets, got %d", IDLength*8, len(rt.buckets))
	}

	for i := 0; i < IDLength*8; i++ {
		if rt.buckets[i] == nil {
			t.Errorf("Expected bucket %d to be initialized", i)
		}
	}
}

func TestAddRoutingTableContact(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)
	contact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")
	rt.AddContact(contact)

	if !rt.IsContactInTable(&contact) {
		t.Errorf("Expected contact %v to be in the table, but it was not.", contact)
	}
}

func TestRemoveRoutingTableContact(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)
	contact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")
	rt.AddContact(contact)

	rt.RemoveContact(contact.ID)

	if rt.IsContactInTable(&contact) {
		t.Errorf("Expected contact %v to be removed from the table, but it was not.", contact)
	}
}

func TestFindClosestContacts(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)
	targetID := NewRandomKademliaID()

	rt.AddContact(NewContact(NewRandomKademliaID(), "localhost:8001"))
	rt.AddContact(NewContact(NewRandomKademliaID(), "localhost:8002"))
	rt.AddContact(NewContact(NewRandomKademliaID(), "localhost:8002"))

	closestContacts := rt.FindClosestContacts(targetID, 2)

	if len(closestContacts) != 2 {
		t.Errorf("Expected 2 closest contacts, got %d", len(closestContacts))
	}
}

func TestGetBucketIndex(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	contact1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")
	contact2 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8001")
	contact3 := NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8001")

	rt.AddContact(contact1)
	rt.AddContact(contact2)
	rt.AddContact(contact3)

	index := rt.getBucketIndex(contact1.ID)
	expected := 159 // contact1 is set to the ID of me, which is expected to return 159 by getBucketIndex

	if index != expected {
		t.Errorf("Expected bucket index %d, but got %d", expected, index)
	}
}

func TestIsContactInTable(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)
	contact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")

	if rt.IsContactInTable(&contact) {
		t.Error("Expected contact to NOT be in the table, but it was.")
	}

	rt.AddContact(contact)

	if !rt.IsContactInTable(&contact) {
		t.Error("Expected contact to be in the table, but it was not.")
	}
}
