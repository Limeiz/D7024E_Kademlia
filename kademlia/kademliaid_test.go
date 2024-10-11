package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"
)

func TestNewKademliaID(t *testing.T) {
	id := NewKademliaID("0000000000000000000000000000000000000001")

	if len(id) != IDLength {
		t.Errorf("Expected length %v, got %v", IDLength, len(id))
	}
}

func TestNewRandomKademliaID(t *testing.T) {
	id1 := NewRandomKademliaID()
	id2 := NewRandomKademliaID()
	if id1.Equals(id2) {
		t.Error("Expected two random KademliaIDs to be different")
	}
	if len(id1) != IDLength {
		t.Errorf("Expected length %v, got %v", IDLength, len(id1))
	}
}

func TestLess(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000001")
	id2 := NewKademliaID("0000000000000000000000000000000000000002")
	if !id1.Less(id2) {
		t.Error("Expected id1 to be less than id2")
	}
	if id2.Less(id1) {
		t.Error("Expected id2 not to be less than id1")
	}
}

func TestEquals(t *testing.T) {
	data := "0000000000000000000000000000000000000001"
	data2 := "0000000000000000000000000000000000000002"
	id1 := NewKademliaID(data)
	id2 := NewKademliaID(data)
	id3 := NewKademliaID(data2)
	if !id1.Equals(id2) {
		t.Error("Expected id1 to be equal to id2")
	}
	if id1.Equals(id3) {
		t.Error("Expected id1 not to be equal to id3")
	}
}

func TestCalcDistance(t *testing.T) {
	id1 := NewKademliaID("0000000000000000000000000000000000000001")
	id2 := NewKademliaID("0000000000000000000000000000000000000002")
	expectedDistance := NewKademliaID("0000000000000000000000000000000000000003")
	distance := id1.CalcDistance(id2)
	if !distance.Equals(expectedDistance) {
		t.Errorf("Expected distance %v, got %v", expectedDistance.String(), distance.String())
	}
}

func TestString(t *testing.T) {
	id := NewKademliaID("0000000000000000000000000000000000000001")
	expected := hex.EncodeToString(id[:])
	if id.String() != expected {
		t.Errorf("Expected string representation %v, got %v", expected, id.String())
	}
}

func TestHashData(t *testing.T) {
	data := "something"
	hashed := HashData(data)
	decoded, err := hex.DecodeString(hashed)
	if err != nil {
		t.Errorf("Expected valid hex string, got error: %v", err)
	}
	if len(decoded) != sha1.Size {
		t.Errorf("Expected hash length of %d bytes, got %d", sha1.Size, len(decoded))
	}
}
