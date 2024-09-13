package kademlia

import (
	"container/list"
)

const k = 20
const alpha = 3
const B = 160

type Kademlia struct {
	routingTable RoutingTable
	hashmap map[KademliaID]string
	kClosestList *list.List
	// me med this nodes address, port
}

type ContactResponse struct {
	nodeID   *KademliaID
	address  string
	port     int
}

func InitKademliaNode() Kademlia{		// add bootstrap node?
	kademlia := &Kademlia{}
	kademlia.hashmap = make(map[KademliaID]string)
	kademlia.routingTable = NewRoutingTable()
	return kademlia	
}


// node1.LookupContact(node2)		 från node1 vill vi returnera en lista med k closest contacts till node2


func (kademlia *Kademlia) LookupContact(target *Contact) { 		// iterativeFindNode
	// If contact doesn't respond, remove from routingTable

	shortlist := list.New()

	var closestNode *Contact	

	visitedNodes := make(map[KademliaID]bool)
 
	closestContacts := kademlia.routingTable.FindClosestContacts(target.ID, alpha)
		
	// for shortlist.Len() < k {
		alphaContacts := []Contact{}
		for _, contact := range closestContacts {
			if !visitedNodes[*contact.ID] {
				alphaContacts = append(alphaContacts, contact)
				visitedNodes[*contact.ID] = true		// maybe move to after sent RPC

				if closestNode == nil || contact.CalcDistance(target.ID) < closestNode.CalcDistance(target.ID){
					closestNode = contact
				}
			}
		}
			// node lookup, SendFindContactMessage
	// }
	
// Less returns true if contact.distance < otherContact.distance
func (contact *Contact) Less(otherContact *Contact) bool {
	return contact.distance.Less(otherContact.distance)
}


	for n, contact := range alphaContacts {
		if n >= alpha {
			break
		}
		// go SendFindContactMessage
		// SendFindContactMessage
		// SendFindDataMessage

		// ch := make(chan string, 3) 
		responseChan := make(chan []ContactResponse, alpha)	

		// ContactCandidates!!!!!




		// Send the first batch of alpha parallel RPCs

		go network.SendFindContactMessage(contact, target, responseChan) // if fail to reply, delete them from shortlist

	
		// Process responses as they come in
		responsesReceived := 0
	
		for responsesReceived < k {
			// Wait for a response from any of the goroutines
			responseList := <-responseChan

	
			// Process the response (update closest node, routing table, etc.)

			for _, response := range responseList {		
				// response.address
				// response.port
				// response.nodeID
				
				newContact := NewContact(response.nodeID, response.address)
				newContact.CalcDistance(kademlia.)

				kademlia.routingTable.AddContact(newContact)

				if responseList != nil && response.nodeID.CalcDistance(target.ID) < closestNode.CalcDistance(target.ID) {
					closestNode = newContact
				}

			}
	
			// Increment the count of responses we've processed
			responsesReceived++

	
			// Once we get a response, send a new request to the next unvisited contact
			for _, contact := range closestContacts {
				if len(alphaContacts) < k && !visitedNodes[*contact.ID] {
					visitedNodes[*contact.ID] = true
					go network.SendFindContactMessage(contact, target, responseChan)
					break
				}
			}
		}
	}
}
	// One contact, return 20 triples, one triple <IP address, Port, NodeID>
	// Uppdatera shortlist



func (kademlia *Kademlia) LookupData(hash string) {  	// iterativeFindValue
	// TODO
	// This is the Kademlia search operation. It is conducted as a node lookup, and so builds a list of k closest contacts. 
	// However, this is done using the FIND_VALUE RPC instead of the FIND_NODE RPC. If at any time during the node lookup the 
	// value is returned instead of a set of contacts, the search is abandoned and the value is returned. 
	// Otherwise, if no value has been found, the list of k closest contacts is returned to the caller.
	// When an iterativeFindValue succeeds, the initiator must store the key/value pair at the closest node seen which did not return the value.

	// variable closestNode that updates (ONLY if the node does NOT have the value)

	// this.RoutingTable.getBucketIndex()

	if (hashmap[hash]!=null){
		return hashmap[hash]
	}
	else{

	}

	

	// if (value is in this node){
	// 	return value						// Store this value in closestNode
	// }
	// else{
	// 	return list of k closest contacts
	// }



	for(KademliaID nodeID in (thisNode)){
		if (hash.Equals(nodeID)){
			return nodeID.value, nodeID
		}
	}
	

	



	// Store(data) in closestNode

}
//string? []byte?
func (kademlia *Kademlia) Store(data []byte) {		//  iterativeStore
	// TODO

	// check size

	keyID := NewKademliaID(data)

	//LookupContact

	//Find k nodes (k=20) and Send value + key there

	
}
