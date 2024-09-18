package kademlia

type Network struct {
}

func Listen(ip string, port int) {
	// TODO
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact, target *Contact, responseChan chan []ContactResponse) { // FIND_NODE Primitive RPC
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) { // FIND_VALUE Primitive RPC
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) { // Primitive STORE RPC
	// TODO
}
