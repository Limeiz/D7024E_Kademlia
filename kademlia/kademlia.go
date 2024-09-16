package kademlia

const k = 20
const alpha = 3

// const B = 160

type Kademlia struct {
	routingTable *RoutingTable
	hashmap      map[KademliaID]string
	// me med this nodes address, port
}

type ContactResponse struct {
	nodeID  *KademliaID
	address string
	port    int
}

func InitKademliaNode(me Contact) *Kademlia { // add bootstrap node?
	kademlia := &Kademlia{
		routingTable: NewRoutingTable(me),
		hashmap:      make(map[KademliaID]string),
	}
	return kademlia
}

// node1.LookupContact(node2)		 från node1 vill vi returnera en lista med k closest contacts till node2

func (kademlia *Kademlia) LookupContact(target *Contact) []Contact { // iterativeFindNode
	// If contact doesn't respond, remove from routingTable

	shortlist := kademlia.routingTable.FindClosestContacts(target.ID, alpha)
	closestNode := shortlist[0] // is this really the closest?
	responseChan := make(chan []ContactResponse, alpha)
	doneChan := make(chan struct{})
	activeRPCs := 0
	candidates := &ContactCandidates{}

	visitedNodes := make(map[KademliaID]bool)

	for len(shortlist) < k {
		alphaContacts := []Contact{} // the nodes that we will send RPCs to
		for _, contact := range shortlist {
			if !visitedNodes[*contact.ID] && len(alphaContacts) < alpha {
				alphaContacts = append(alphaContacts, contact)
				visitedNodes[*contact.ID] = true
				if contact.Less(&closestNode) {
					closestNode = contact
				}
			}
		}

		for _, contact := range alphaContacts {
			// ContactCandidates?

			// Send the first batch of alpha parallel RPCs
			activeRPCs++
			go network.SendFindContactMessage(contact, target, responseChan) // if fail to reply, delete them from shortlist, maybe move all this to a func

		}

		go processResponses(kademlia, responseChan, &shortlist, visitedNodes, doneChan, closestNode, activeRPCs, *target, candidates) //ugly

		<-doneChan
	}

	return shortlist
}

func processResponses(kademlia *Kademlia, responseChan <-chan []ContactResponse, shortlist *[]Contact, visitedNodes map[KademliaID]bool, doneChan chan struct{}, closestNode Contact, activeRPCs int, target Contact, candidates *ContactCandidates) {
	for len(*shortlist) < k && activeRPCs > 0 {
		select {
		case responseList := <-responseChan:
			activeRPCs--

			for _, response := range responseList {
				newContact := NewContact(response.nodeID, response.address)
				newContact.CalcDistance(target.ID)
				kademlia.routingTable.AddContact(newContact)

				if len(*shortlist) < k {
					*shortlist = append(*shortlist, newContact) // CHANGE HERE, USE CANDIDATES, should our whole shortlist be candidates instead?
				}

				if newContact.Less(&closestNode) {
					closestNode = newContact
				}

				// If we haven't visited this contact, send another RPC
				if !visitedNodes[*newContact.ID] {
					visitedNodes[*newContact.ID] = true
					activeRPCs++
					go network.SendFindContactMessage(newContact, target, responseChan)
				}
			}

		default:
			// Search is finished
			if activeRPCs == 0 {
				doneChan <- struct{}{}
				return
			}
		}
	}
}

// 		// Process responses as they come in
// 		responsesReceived := 0

// 		for responsesReceived < k {
// 			// Wait for a response from any of the goroutines
// 			responseList := <-responseChan

// 			// Process the response (update closest node, routing table, etc.)

// 			for _, response := range responseList {

// 				newContact := NewContact(response.nodeID, response.address)
// 				newContact.CalcDistance(kademlia.)

// 				kademlia.routingTable.AddContact(newContact)

// 				if responseList != nil && response.nodeID.CalcDistance(target.ID) < closestNode.CalcDistance(target.ID) {
// 					closestNode = newContact
// 				}

// 			}

// 			// Increment the count of responses we've processed
// 			responsesReceived++

// 			// Once we get a response, send a new request to the next unvisited contact
// 			for _, contact := range closestContacts {
// 				if len(alphaContacts) < k && !visitedNodes[*contact.ID] {
// 					visitedNodes[*contact.ID] = true
// 					go network.SendFindContactMessage(contact, target, responseChan)
// 					break
// 				}
// 			}
// 		}
// 	}
// }
// One contact, return 20 triples, one triple <IP address, Port, NodeID>
// Uppdatera shortlist

func (kademlia *Kademlia) LookupData(hash string) { // iterativeFindValue
	// TODO
	// This is the Kademlia search operation. It is conducted as a node lookup, and so builds a list of k closest contacts.
	// However, this is done using the FIND_VALUE RPC instead of the FIND_NODE RPC. If at any time during the node lookup the
	// value is returned instead of a set of contacts, the search is abandoned and the value is returned.
	// Otherwise, if no value has been found, the list of k closest contacts is returned to the caller.
	// When an iterativeFindValue succeeds, the initiator must store the key/value pair at the closest node seen which did not return the value.

	// variable closestNode that updates (ONLY if the node does NOT have the value)

	// this.RoutingTable.getBucketIndex()

	// if hashmap[hash] != null {
	// 	return hashmap[hash]
	// }
	// else{

	// }

	// if (value is in this node){
	// 	return value						// Store this value in closestNode
	// }
	// else{
	// 	return list of k closest contacts
	// }

	// for(KademliaID nodeID in (thisNode)){
	// 	if (hash.Equals(nodeID)){
	// 		return nodeID.value, nodeID
	// 	}
	// }

	// Store(data) in closestNode

}

// string? []byte?
func (kademlia *Kademlia) Store(data []byte) { //  iterativeStore
	// TODO

	// check size

	// keyID := NewKademliaID(data)

	//LookupContact

	//Find k nodes (k=20) and Send value + key there

}
