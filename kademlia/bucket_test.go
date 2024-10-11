package kademlia

import (
	"testing"
)

func TestNewBucket(t *testing.T) {
	b := newBucket()
	if b == nil {
		t.Fatalf("Expected new bucket to be non-nil")
	}
	if b.Len() != 0 {
		t.Fatalf("Expected new bucket length to be 0, got %d", b.Len())
	}
}

func TestBucketAddContact(t *testing.T) {
	bucket := newBucket()
	id := NewRandomKademliaID()
	contact := NewContact(id, "192.168.0.1")

	bucket.AddContact(contact)
	if bucket.Len() != 1 {
		t.Errorf("Expected bucket length to be 1, got %d", bucket.Len())
	}
}

func TestRemoveContact(t *testing.T) {
	bucket := newBucket()
	id := NewRandomKademliaID()
	contact := NewContact(id, "192.168.0.1")

	bucket.AddContact(contact)

	bucket.RemoveContact(contact.ID)
	if bucket.Len() != 0 {
		t.Errorf("Expected bucket length to be 0 after removing contact, got %d", bucket.Len())
	}
}

func fillBucket(bucket *bucket) *bucket {
	for i := 0; i < bucketSize*2; i++ {
		contact := NewContact(NewRandomKademliaID(), "192.168.0.1")
		bucket.AddContact(contact)
	}
	return bucket
}

func TestBucketGetContact(t *testing.T) {
	bucket := newBucket()
	bucket = fillBucket(bucket)
	targetID := NewRandomKademliaID()

	contacts := bucket.GetContactAndCalcDistance(targetID)

	if contacts[0].distance == nil {
		t.Errorf("Expected first contact's distance to be non-nil, but got nil")
	}

	if contacts[1].distance == nil {
		t.Errorf("Expected second contact's distance to be non-nil, but got nil")
	}
}

func TestLen(t *testing.T) {
	bucket := newBucket()

	if got := bucket.Len(); got != 0 {
		t.Errorf("Expected bucket length to be 0, got %d", got)
	}

	bucket = fillBucket(bucket)

	if got := bucket.Len(); got != 20 {
		t.Errorf("Expected bucket length to be 20, got %d", got)
	}
}
