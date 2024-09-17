package kademlia

import (
	"log"
	"net"
)

type Network struct {
}

type MessageType int32

const (
	PING MessageType = 0
	PONG MessageType = 1
)

type MessageHeader struct {
	HeaderLength  int32
	HeaderTag     [4]byte
	Type          MessageType
	SenderIP      net.IP
	ReceiverIP    net.IP
	ContentOffset int32
}

// func Listen(ip string, port int) {

// }

func OpenPortAndListen(port int) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	connection, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	buffer := make([]byte, 1024) // Buffer to store incoming data
	for {
		// Read UDP packets
		n, remoteAddr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading UDP packet:", err)
			continue
		}

		log.Printf("Received %d bytes from %s: %s\n", n, remoteAddr, string(buffer[:n]))

		_, err = connection.WriteToUDP([]byte("pong"), remoteAddr)
		if err != nil {
			log.Println("Error sending response:", err)
		}
	}
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
