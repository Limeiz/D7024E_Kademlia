package kademlia

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Network struct {
	Node             *Kademlia
	ResponseMap      map[uint64]chan MessageData
	ResponseMapMutex sync.Mutex
	Port             int
}

type MessageType uint8
type MessageDirection uint8

const (
	PING MessageType = 0
	PONG MessageType = 1
)

const (
	REQUEST  MessageDirection = 0
	RESPONSE MessageDirection = 1
)

type MessageHeader struct {
	HeaderLength uint32
	HeaderTag    [4]byte
	Direction    MessageDirection
	Type         MessageType
	MessageID    uint64
	SenderID     KademliaID
	RecieverID   KademliaID
	BodyLength   uint32
	SenderIP     string
	ReceiverIP   string
	ReceiverZone string // The docker container name, since it doesn't count as a IP
}

type MessageData struct {
	Header MessageHeader
	Data   []byte
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

func MessageDirectionToString(message_direction MessageDirection) string {
	switch message_direction {
	case REQUEST:
		return "REQUEST"
	case RESPONSE:
		return "RESPONSE"
	default:
		return "NIL"
	}
}

func SerializeMessage(messageData MessageData) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(messageData)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DeserializeMessage(data []byte) (MessageData, error) {
	var messageData MessageData
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&messageData)
	if err != nil {
		return messageData, err
	}
	return messageData, nil
}

func InitNetwork(node *Kademlia) *Network {
	network := &Network{
		Node:        node,
		ResponseMap: make(map[uint64]chan MessageData),
		Port:        8080,
	}
	return network
}

func (network *Network) ServerInit() {
	http.HandleFunc("/", network.DefaultController)  // Corrected net.http to net/http
	http.HandleFunc("/ping", network.PingController) // Corrected net.http to net/http
	http.HandleFunc("/getid", network.GetID)
}

func (network *Network) ServerStart(port int) {
	network.Port = port
	addr := net.TCPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	err := http.ListenAndServe(addr.String(), nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server closing \n")
	} else if err != nil {
		log.Printf("Server error: %s\n", err)
	}
}

func (network *Network) HandleMessages(buffer []byte, n int, addr *net.UDPAddr) {
	message, err := DeserializeMessage(buffer[:n])
	if err != nil {
		log.Printf("Failed to deserialize message from %s: %v\n", addr, err)
		return
	}
	log.Printf("Received a %s %s message from %s\n", MessageDirectionToString(message.Header.Direction), MessageTypeToString(message.Header.Type), addr)

	if message.Header.Direction == RESPONSE {
		responseChan, exists := network.ResponseMap[message.Header.MessageID]
		if exists {
			responseChan <- message
			delete(network.ResponseMap, message.Header.MessageID)
		} else {
			log.Printf("No waiting request for message ID %d from %s\n", message.Header.MessageID, addr)
		}
	} else {
		// Handle request messages (e.g., PING) here
		if message.Header.Type == PING {
			new_contact := NewContact(&message.Header.SenderID, message.Header.SenderIP)
			err := network.SendMessage(&new_contact, PONG, RESPONSE, nil, message.Header.MessageID)
			if err != nil {
				log.Printf("Failed to send PONG to %s: %v", new_contact.Address, err)
			}
		}
	}
}

func (network *Network) OpenPortAndListen(port int) {
	network.Port = port
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	connection, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading UDP packet:", err)
			continue
		}
		bufferCopy := make([]byte, n)
		copy(bufferCopy, buffer[:n])
		go network.HandleMessages(bufferCopy, n, remoteAddr)
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

func GenerateMessageID() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

func (network *Network) SendMessage(contact *Contact, messageType MessageType, messageDir MessageDirection, data []byte, message_id ...uint64) error {
	var messageID uint64
	if len(message_id) > 0 {
		messageID = message_id[0]
	} else {
		messageID = GenerateMessageID()
	}
	messageHeader := MessageHeader{
		HeaderLength: uint32(binary.Size(MessageHeader{})),
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    messageDir,
		Type:         messageType,
		MessageID:    messageID,
		SenderID:     *network.Node.Routes.Me.ID,
		RecieverID:   *contact.ID,
		BodyLength:   uint32(len(data)),
		SenderIP:     GetLocalIP(),
		ReceiverIP:   contact.Address,
	}
	messageData := MessageData{
		Header: messageHeader,
		Data:   data,
	}
	messageBytes, err := SerializeMessage(messageData)
	if err != nil {
		return err
	}
	conn, err := net.Dial("udp", contact.Address+":"+strconv.Itoa(network.Port))
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("Writing message of type %s to %s \n", MessageTypeToString(messageData.Header.Type), messageData.Header.ReceiverIP)
	_, err = conn.Write(messageBytes)
	return err
}

func (network *Network) SendMessageAndWait(contact *Contact, messageType MessageType, messageDir MessageDirection, data []byte, message_id ...uint64) (MessageData, error) {
	var messageID uint64
	if len(message_id) > 0 {
		messageID = message_id[0]
	} else {
		messageID = GenerateMessageID()
	}
	responseChan := make(chan MessageData)
	network.ResponseMapMutex.Lock()
	network.ResponseMap[messageID] = responseChan
	network.ResponseMapMutex.Unlock()
	err := network.SendMessage(contact, messageType, messageDir, data, messageID)
	if err != nil {
		return MessageData{}, err
	}
	select {
	case response := <-responseChan:
		network.ResponseMapMutex.Lock()
		delete(network.ResponseMap, messageID)
		network.ResponseMapMutex.Unlock()
		return response, nil
	case <-time.After(5 * time.Second):
		network.ResponseMapMutex.Lock()
		delete(network.ResponseMap, messageID)
		network.ResponseMapMutex.Unlock()
		return MessageData{}, errors.New("timeout waiting for response")
	}
}

func (network *Network) SendMessageAndWaitByIP(addr string, messageType MessageType, messageDir MessageDirection, data []byte) (MessageData, error) {
	id := NewRandomKademliaID()
	contact := NewContact(id, addr)
	return network.SendMessageAndWait(&contact, messageType, messageDir, data)
}

func (network *Network) SendMessageByIP(addr string, messageType MessageType, messageDir MessageDirection, data []byte) error {
	id := NewRandomKademliaID()
	contact := NewContact(id, addr)
	return network.SendMessage(&contact, messageType, messageDir, data)
}
