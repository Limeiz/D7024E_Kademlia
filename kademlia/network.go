package kademlia

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http" // Corrected import
	"os"
	"time"
	"errors"
)

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

func MessageTypeToString(message_type MessageType) string {
	switch message_type {
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	default:
		return "NIL"
	}
}

func SerializeMessage(header MessageHeader) ([]byte, error) {
	buf := make([]byte, 32) // Arbitrary buffer size
	binary.LittleEndian.PutUint32(buf[0:], uint32(header.HeaderLength))
	copy(buf[4:], header.HeaderTag[:])
	binary.LittleEndian.PutUint32(buf[8:], uint32(header.Type))
	copy(buf[12:], header.SenderIP.To4())
	copy(buf[16:], header.ReceiverIP.To4())
	binary.LittleEndian.PutUint32(buf[20:], uint32(header.ContentOffset))

	return buf, nil
}

func DeserializeMessage(data []byte) (MessageHeader, error) {
	var header MessageHeader
	if len(data) < 32 { // Ensure the message is at least the size of the MessageHeader
		return header, fmt.Errorf("message too short")
	}

	header.HeaderLength = int32(binary.LittleEndian.Uint32(data[0:4]))
	copy(header.HeaderTag[:], data[4:8])
	header.Type = MessageType(binary.LittleEndian.Uint32(data[8:12]))
	header.SenderIP = net.IPv4(data[12], data[13], data[14], data[15])
	header.ReceiverIP = net.IPv4(data[16], data[17], data[18], data[19])
	header.ContentOffset = int32(binary.LittleEndian.Uint32(data[20:24]))

	return header, nil
}

func ServerInit() {
	http.HandleFunc("/", DefaultController)    // Corrected net.http to net/http
	http.HandleFunc("/ping", PingController)   // Corrected net.http to net/http
}

func ServerStart(port int) {
	addr := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	err := http.ListenAndServe(addr.String(), nil) // Pass addr as a string instead of &addr

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server closing \n")
	} else if err != nil {
		log.Printf("Server error: %s\n", err)
	}
}

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

		message, err := DeserializeMessage(buffer[:n])
		if err != nil {
			log.Printf("Failed to deserialize message from %s: %v\n", remoteAddr, err)
			continue
		}

		// Process the message
		log.Printf("Received %s message from %s\n", MessageTypeToString(message.Type), remoteAddr)

		if message.Type == PING {
			log.Printf("Received PING from %s\n", remoteAddr)
		}
	}
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func SendPingMessage(contact *Contact) {
	conn, err := net.DialTimeout("udp", contact.Address+":"+os.Getenv("COMMUNICATION_PORT"), 2*time.Second)
	if err != nil {
		log.Printf("Failed to connect to %s: %v", contact.Address, err)
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	message := MessageHeader{
		HeaderLength:  int32(binary.Size(MessageHeader{})),
		HeaderTag:     [4]byte{'K', 'A', 'D', 'M'},
		Type:          PING,
		SenderIP:      localAddr.IP,
		ReceiverIP:    net.ParseIP(contact.Address),
		ContentOffset: 0,
	}

	// Serialize the message to bytes
	messageBytes, err := SerializeMessage(message)
	if err != nil {
		log.Printf("Failed to serialize message: %v", err)
		return
	}
	_, err = conn.Write(messageBytes)
	if err != nil {
		log.Printf("Failed to send Ping to %s: %v", contact.Address, err)
		return
	}

	log.Printf("Ping sent to %s", contact.Address)

	conn.Close()
}

func SendPingMessageByIP(ip_address string) {
	id := NewRandomKademliaID()
	contact := NewContact(id, ip_address)
	SendPingMessage(&contact)
}

func SendMessage(contact *Contact, message_type MessageType) {
	// TODO
}

func SendFindContactMessage(contact *Contact) {
	// TODO
}

func SendFindDataMessage(hash string) {
	// TODO
}

func SendStoreMessage(data []byte) {
	// TODO
}
