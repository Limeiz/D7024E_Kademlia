package kademlia

import (
	"fmt"
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

	rt.buckets[0] = nil

	if rt.IsContactInTable(&contact) {
		t.Error("Expected contact to NOT be in the table, but it was.")
	}

	rt.AddContact(contact)

	if !rt.IsContactInTable(&contact) {
		t.Error("Expected contact to be in the table, but it was not.")
	}
}

func TestRoutingTableString(t *testing.T) {
	me := NewContact(NewKademliaID("ffffffffffffffffffffffffffffffffffffffff"), "localhost:8000")
	rt := NewRoutingTable(me)

	contact1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001")
	contact2 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	rt.AddContact(contact1)
	rt.AddContact(contact2)
	rt.AddContact(NewContact(NewRandomKademliaID(), "localhost:8001"))
	rt.AddContact(NewContact(NewRandomKademliaID(), "localhost:8003"))

	fmt.Println(rt.String())
}

// func TestRoutingTableString(t *testing.T) {
// 	// Create a new routing table with a contact
// 	me := NewContact(NewKademliaID("ffffffffffffffffffffffffffffffffffffffff"), "localhost:8000")
// 	rt := NewRoutingTable(me)

// 	// Check the string representation for an empty routing table
// 	expectedEmpty := ""
// 	for i := 0; i < IDLength*8; i++ {
// 		expectedEmpty += fmt.Sprintf("B%d: {}\n", i)
// 	}

// 	if result := rt.String(); result != expectedEmpty {
// 		t.Errorf("Expected empty string representation:\n%q\nbut got:\n%q", expectedEmpty, result)
// 	}

// 	// Add a contact to the routing table
// 	contact1 := NewContact(NewKademliaID("fffffffffffffffffffffffffffffffffffffffe"), "localhost:8001")
// 	rt.AddContact(contact1)

// 	// Check the string representation after adding contact
// 	expectedNonEmpty := fmt.Sprintf(
// 		"B0: {(fffffffffffffffffffffffffffffffffffffffe, localhost:8001)}\n" +
// 			"B1: {}\n" +
// 			"B2: {}\nB3: {}\nB4: {}\nB5: {}\nB6: {}\nB7: {}\nB8: {}\nB9: {}\n" +
// 			"B10: {}\nB11: {}\nB12: {}\nB13: {}\nB14: {}\nB15: {}\nB16: {}\n" +
// 			"B17: {}\nB18: {}\nB19: {}\nB20: {}\nB21: {}\nB22: {}\nB23: {}\n" +
// 			"B24: {}\nB25: {}\nB26: {}\nB27: {}\nB28: {}\nB29: {}\nB30: {}\n" +
// 			"B31: {}\nB32: {}\nB33: {}\nB34: {}\nB35: {}\nB36: {}\nB37: {}\n" +
// 			"B38: {}\nB39: {}\nB40: {}\nB41: {}\nB42: {}\nB43: {}\nB44: {}\n" +
// 			"B45: {}\nB46: {}\nB47: {}\nB48: {}\nB49: {}\nB50: {}\nB51: {}\n" +
// 			"B52: {}\nB53: {}\nB54: {}\nB55: {}\nB56: {}\nB57: {}\nB58: {}\n" +
// 			"B59: {}\nB60: {}\nB61: {}\nB62: {}\nB63: {}\nB64: {}\nB65: {}\n" +
// 			"B66: {}\nB67: {}\nB68: {}\nB69: {}\nB70: {}\nB71: {}\nB72: {}\n" +
// 			"B73: {}\nB74: {}\nB75: {}\nB76: {}\nB77: {}\nB78: {}\nB79: {}\n" +
// 			"B80: {}\nB81: {}\nB82: {}\nB83: {}\nB84: {}\nB85: {}\nB86: {}\n" +
// 			"B87: {}\nB88: {}\nB89: {}\nB90: {}\nB91: {}\nB92: {}\nB93: {}\n" +
// 			"B94: {}\nB95: {}\nB96: {}\nB97: {}\nB98: {}\nB99: {}\nB100: {}\n" +
// 			"B101: {}\nB102: {}\nB103: {}\nB104: {}\nB105: {}\nB106: {}\n" +
// 			"B107: {}\nB108: {}\nB109: {}\nB110: {}\nB111: {}\nB112: {}\n" +
// 			"B113: {}\nB114: {}\nB115: {}\nB116: {}\nB117: {}\nB118: {}\n" +
// 			"B119: {}\nB120: {}\nB121: {}\nB122: {}\nB123: {}\nB124: {}\n" +
// 			"B125: {}\nB126: {}\nB127: {}\nB128: {}\nB129: {}\nB130: {}\n" +
// 			"B131: {}\nB132: {}\nB133: {}\nB134: {}\nB135: {}\nB136: {}\n" +
// 			"B137: {}\nB138: {}\nB139: {}\nB140: {}\nB141: {}\nB142: {}\n" +
// 			"B143: {}\nB144: {}\nB145: {}\nB146: {}\nB147: {}\nB148: {}\n" +
// 			"B149: {}\nB150: {}\nB151: {}\nB152: {}\nB153: {}\nB154: {}\n" +
// 			"B155: {}\nB156: {}\nB157: {}\nB158: {}\nB159: {}\n",
// 	)

// 	// Update the string representation after adding contact
// 	if result := rt.String(); result != expectedNonEmpty {
// 		t.Errorf("Expected string representation:\n%q\nbut got:\n%q", expectedNonEmpty, result)
// 	}
// }
