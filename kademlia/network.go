package kademlia

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
	"io"
	"context"
)

type Network struct {
	Node             *Kademlia
	ResponseMap      map[KademliaID]chan MessageData
	ResponseMapMutex sync.Mutex
	Port             int
	Timeout 		 int
}

type MessageType uint8
type MessageDirection uint8

const (
	PING       MessageType = 0
	FIND_NODE  MessageType = 1
	FIND_VALUE MessageType = 2
	STORE      MessageType = 3
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
	SenderIP     [15]byte
	ReceiverIP   [15]byte
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

func EncodeMessageHeader(header MessageHeader) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, header.HeaderLength); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.HeaderTag); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.Direction); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.MessageID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.SenderID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.ReceiverID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.BodyLength); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.SenderIP); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, header.ReceiverIP); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeMessageHeader(data []byte) (MessageHeader, error) {
	buf := bytes.NewReader(data)
	var header MessageHeader
	if err := binary.Read(buf, binary.BigEndian, &header.HeaderLength); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.HeaderTag); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Direction); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Type); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.MessageID); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.SenderID); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.ReceiverID); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.BodyLength); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.SenderIP); err != nil {
		return header, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.ReceiverIP); err != nil {
		return header, err
	}

	return header, nil
}

func SerializeMessage(messageData MessageData) ([]byte, error) {
	headerBytes, err := EncodeMessageHeader(messageData.Header)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.Write(headerBytes)
	buf.Write(messageData.Data) // Data is already in bytes

	return buf.Bytes(), nil
}

func DeserializeMessage(data []byte) (MessageData, error) {
	var msg MessageData
	buf := bytes.NewReader(data)
	headerBytes := make([]byte, binary.Size(MessageHeader{}))
	if _, err := io.ReadFull(buf, headerBytes); err != nil {
		return msg, err
	}
	header, err := DecodeMessageHeader(headerBytes)
	if err != nil {
		return msg, err
	}
	msg.Header = header
	msg.Data = make([]byte, msg.Header.BodyLength)
	if _, err := io.ReadFull(buf, msg.Data); err != nil {
		return msg, err
	}

	return msg, nil
}

func InitNetwork(node *Kademlia, timeout int) *Network {
	network := &Network{
		Node:        node,
		ResponseMap: make(map[KademliaID]chan MessageData),
		Port:        8080,
		Timeout: 	 timeout,
	}
	return network
}

func (network *Network) ServerInit() {
	http.HandleFunc("/", network.DefaultController)
	http.HandleFunc("/ping", network.PingController)
	http.HandleFunc("/put", network.PutController)
	http.HandleFunc("/get", network.GetController)
	http.HandleFunc("/getid", network.GetID)
	http.HandleFunc("/show-routing-table", network.ShowRoutingTableController)
	http.HandleFunc("/show-storage", network.ShowStorageController)
	http.HandleFunc("/exit", network.ExitController)
}

func (network *Network) ServerStart(port int) {
    network.Port = port
    addr := net.TCPAddr{
        Port: port,
        IP:   net.ParseIP("0.0.0.0"),
    }

    server := &http.Server{
        Addr: addr.String(),
    }

    errChan := make(chan error, 1)

    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    select {
    case <-network.Node.ShutdownChan:
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := server.Shutdown(ctx); err != nil {
            log.Printf("Server Shutdown Failed: %+v", err)
        } else {
            log.Printf("Server Shutdown gracefully")
        }
    case err := <-errChan:
        if err != nil {
            log.Printf("Server error: %s\n", err)
        }
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
		network.ResponseMapMutex.Lock()
		responseChan, exists := network.ResponseMap[message.Header.MessageID]
		network.ResponseMapMutex.Unlock()
		if exists {
			responseChan <- message
		} else {
			log.Printf("No waiting request for message[%s] from %s\n", message.Header.MessageID.String(), addr)
		}
	} else {
		sender_contact := NewContact(&message.Header.SenderID, Byte15ArrayToString(message.Header.SenderIP))
		// Handle request messages (e.g., PING) here
		if message.Header.Type == PING {
			network.Node.Pong(&sender_contact, &message.Header.MessageID)
		} else if message.Header.Type == FIND_NODE {
			responseData, err := network.Node.ProcessFindContactMessage(&message.Data, sender_contact)
			if err != nil {
				log.Printf("Failed to process FIND_NODE message %s: %v \n", sender_contact.Address, err)
			}
			err = network.SendMessage(&sender_contact, FIND_NODE, RESPONSE, responseData, &message.Header.MessageID)
			if err != nil {
				log.Printf("Failed to send FIND_NODE to %s: %v \n", sender_contact.Address, err)
			}
		} else if message.Header.Type == FIND_VALUE {
			responseData, err := network.Node.ProcessFindValueMessage(&message.Data)
			if err != nil {
				log.Printf("Failed to process FIND_VALUE message %s: %v \n", sender_contact.Address, err)
			}
			err = network.SendMessage(&sender_contact, FIND_VALUE, RESPONSE, responseData, &message.Header.MessageID)
			if err != nil {
				log.Printf("Failed to send FIND_VALUE to %s: %v \n", sender_contact.Address, err)
			}
		} else if message.Header.Type == STORE {
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

    connection.SetReadBuffer(65535)

    buffer := make([]byte, 65535)
    done := make(chan struct{})
    go func() {
        <-network.Node.ShutdownChan
        connection.Close()
        close(done)
    }()

    for {
        n, remoteAddr, err := connection.ReadFromUDP(buffer)
        if err != nil {
            select {
            case <-done:
                log.Println("UDP listener shutting down gracefully")
                return
            default:
                log.Println("Error reading UDP packet:", err)
                continue
            }
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

func StringTo15ByteArray(str string) [15]byte {
	var arr [15]byte
	copy(arr[:], str)
	return arr
}

func Byte15ArrayToString(buffer [15]byte) string{
	return string(bytes.Trim(buffer[:], "\x00"))
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
		SenderIP:     StringTo15ByteArray(GetLocalIP()),
		ReceiverIP:   StringTo15ByteArray(contact.Address),
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
	_, exists := network.ResponseMap[*messageID]
	if exists{
		log.Printf("Warning: The messageID %s already exists", messageID.String())
		return MessageData{}, errors.New("MessageID already has a channel associated with it")
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

	case <-time.After(time.Duration(network.Timeout) * time.Second):
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
