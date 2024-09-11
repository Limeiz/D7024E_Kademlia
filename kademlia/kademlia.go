package kademlia

type Kademlia struct {
	routingTable RoutingTable
	hashmap map[KademliaID]string
	// me med this nodes adress,port
}

func InitKademliaNode() Kademlia{
	this.hashmap := make(map[KademliaID]string)
	this.routingTable := NewRoutingTable()
	
}

func (kademlia *Kademlia) LookupContact(target *Contact) { 		// iterativeFindNode
	// TODO

	// If contact doesn't respond, remove from routingTable



}

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
