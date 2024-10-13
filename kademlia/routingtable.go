package kademlia

import (
	"fmt"
)

const bucketSize = 20

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	Me      Contact
	buckets [IDLength * 8]*bucket
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.Me = me
	return routingTable
}

// AddContact add a new contact to the correct Bucket, ONLY if not Me or already in it
func (routingTable *RoutingTable) AddContact(contact Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	if contact != routingTable.Me || routingTable.IsContactInTable(&contact) {
		bucket.AddContact(contact)
	}

}

// RemoveContact remove a contact from the correct Bucket by ID
func (routingTable *RoutingTable) RemoveContact(contact_id *KademliaID) {
	bucketIndex := routingTable.getBucketIndex(contact_id)
	bucket := routingTable.buckets[bucketIndex]
	bucket.RemoveContact(contact_id)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.Me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}

func (routingTable *RoutingTable) String() string {
	var result string
	for i, bucket := range routingTable.buckets {
		result += fmt.Sprintf("B%d: {", i)
		for e := bucket.list.Front(); e != nil; e = e.Next() {
			contact := e.Value.(Contact)
			result += fmt.Sprintf("(%s, %s)", contact.ID.String(), contact.Address)
			if e.Next() != nil {
				result += ", "
			}
		}
		result += "}\n"
	}
	return result
}

// Check if a given contact is already in the routing table
func (routingTable *RoutingTable) IsContactInTable(contact *Contact) bool {
	for _, bucket := range routingTable.buckets {
		if bucket == nil {
			continue
		}

		for e := bucket.list.Front(); e != nil; e = e.Next() {
			if e.Value.(Contact).ID.Equals(contact.ID) {
				return true
			}
		}
	}
	return false
}
