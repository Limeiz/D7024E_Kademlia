package kademlia

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Network struct {
	Node             *Kademlia
	ResponseMap      map[KademliaID]chan MessageData
	ResponseMapMutex sync.Mutex
	Port             int
}

type MessageType uint8
type MessageDirection uint8

const (
	PING            MessageType = 0
	FIND_NODE       MessageType = 1
	FIND_VALUE      MessageType = 2
	RETURN_VALUE    MessageType = 3 // This isn't really needed since a return value message can be defined as <RESPONSE, FIND_VALUE>
	RETURN_CONTACTS MessageType = 4 // Same as above
	STORE			MessageType = 5
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
	MessageID    KademliaID
	SenderID     KademliaID
	ReceiverID   KademliaID
	BodyLength   uint32
	SenderIP     string
	ReceiverIP   string
}

type MessageData struct {
	Header MessageHeader
	Data   []byte
}

func MessageTypeToString(message_type MessageType) string {
	switch message_type {
	case PING:
		return "PING"
	case FIND_NODE:
		return "FIND_NODE"
	case FIND_VALUE:
		return "FIND_VALUE"
	case STORE:
		return "STORE"
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
		ResponseMap: make(map[KademliaID]chan MessageData),
		Port:        8080,
	}
	return network
}

func (network *Network) ServerInit() {
	http.HandleFunc("/", network.DefaultController)  // Corrected net.http to net/http
	http.HandleFunc("/ping", network.PingController) // Corrected net.http to net/http
	http.HandleFunc("/getid", network.GetID)
	http.HandleFunc("/show-routing-table", network.ShowRoutingTableController)
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
	log.Printf("Received a <%s,%s> message[%s] from %s\n", MessageDirectionToString(message.Header.Direction), MessageTypeToString(message.Header.Type), message.Header.MessageID.String(), addr)

	if message.Header.Direction == RESPONSE {
		responseChan, exists := network.ResponseMap[message.Header.MessageID]
		if exists {
			responseChan <- message
		} else {
			log.Printf("No waiting request for message[%s] from %s\n", message.Header.MessageID.String(), addr)
		}
	} else {
		sender_contact := NewContact(&message.Header.SenderID, message.Header.SenderIP)
		// Handle request messages (e.g., PING) here
		if message.Header.Type == PING {
			network.Node.Pong(&sender_contact, &message.Header.MessageID)
		}else if message.Header.Type == FIND_NODE {
			responseData, err := network.Node.ProcessFindContactMessage(&message.Data, sender_contact)
			if err != nil {
				log.Printf("Failed to process FIND_NODE message %s: %v", sender_contact.Address, err)
			}
			err = network.SendMessage(&sender_contact, FIND_NODE, RESPONSE, responseData, &message.Header.MessageID)
			if err != nil {
				log.Printf("Failed to send FIND_NODE to %s: %v", sender_contact.Address, err)
			}
		}else if message.Header.Type == FIND_VALUE {
			responseData, msgType, err := network.Node.ProcessFindValueMessage(&message.Data) //msgType to signal if it is value or contacts
			if err != nil {
				log.Printf("Failed to process FIND_VALUE message %s: %v", sender_contact.Address, err)
			}
			err = network.SendMessage(&sender_contact, msgType, RESPONSE, responseData, &message.Header.MessageID)
			if err != nil {
				log.Printf("Failed to send FIND_VALUE to %s: %v", sender_contact.Address, err)
			}
		}else if message.Header.Type == STORE{
			network.Node.RecieveStoreRPC(&message.Data)
			
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

func (network *Network) SendMessage(contact *Contact, messageType MessageType, messageDir MessageDirection, data []byte, message_id ...*KademliaID) error {
	var messageID *KademliaID
	if len(message_id) > 0 {
		messageID = message_id[0]
	} else {
		messageID = NewRandomKademliaID()
	}

	messageHeader := MessageHeader{
		HeaderLength: uint32(binary.Size(MessageHeader{})),
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    messageDir,
		Type:         messageType,
		MessageID:    *messageID,
		SenderID:     *network.Node.Routes.Me.ID,
		ReceiverID:   *contact.ID,
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

	log.Printf("Writing message[%s] of type <%s,%s> to %s \n", messageData.Header.MessageID.String(), MessageDirectionToString(messageData.Header.Direction), MessageTypeToString(messageData.Header.Type), messageData.Header.ReceiverIP)
	_, err = conn.Write(messageBytes)
	return err
}

func (network *Network) SendMessageAndWait(contact *Contact, messageType MessageType, messageDir MessageDirection, data []byte, message_id ...*KademliaID) (MessageData, error) {
	var messageID *KademliaID
	if len(message_id) > 0 {
		messageID = message_id[0]
	} else {
		messageID = NewRandomKademliaID()
	}

	responseChan := make(chan MessageData)
	network.ResponseMapMutex.Lock()
	network.ResponseMap[*messageID] = responseChan
	network.ResponseMapMutex.Unlock()

	err := network.SendMessage(contact, messageType, messageDir, data, messageID)
	if err != nil {
		return MessageData{}, err
	}

	select {
	case response := <-responseChan:
		network.ResponseMapMutex.Lock()
		delete(network.ResponseMap, *messageID)
		network.ResponseMapMutex.Unlock()
		return response, nil

	case <-time.After(5 * time.Second):
		network.ResponseMapMutex.Lock()
		delete(network.ResponseMap, *messageID)
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
