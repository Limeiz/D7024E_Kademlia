package kademlia

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestMessageTypeToString(t *testing.T) {
	result := MessageTypeToString(PING)
	expected := "PING"
	if result != expected {
		t.Errorf("Expected %s for PING, got %s", expected, result)
	}

	result = MessageTypeToString(FIND_NODE)
	expected = "FIND_NODE"
	if result != expected {
		t.Errorf("Expected %s for FIND_NODE, got %s", expected, result)
	}

	result = MessageTypeToString(FIND_VALUE)
	expected = "FIND_VALUE"
	if result != expected {
		t.Errorf("Expected %s for FIND_VALUE, got %s", expected, result)
	}

	result = MessageTypeToString(STORE)
	expected = "STORE"
	if result != expected {
		t.Errorf("Expected %s for STORE, got %s", expected, result)
	}

	result = MessageTypeToString(MessageType(99))
	expected = "NIL"
	if result != expected {
		t.Errorf("Expected %s for unknown MessageType, got %s", expected, result)
	}
}

func TestMessageDirectionToString(t *testing.T) {
	result := MessageDirectionToString(REQUEST)
	expected := "REQUEST"
	if result != expected {
		t.Errorf("Expected %s for REQUEST, got %s", expected, result)
	}

	result = MessageDirectionToString(RESPONSE)
	expected = "RESPONSE"
	if result != expected {
		t.Errorf("Expected %s for RESPONSE, got %s", expected, result)
	}

	result = MessageDirectionToString(MessageDirection(99))
	expected = "NIL"
	if result != expected {
		t.Errorf("Expected %s for unknown MessageDirection, got %s", expected, result)
	}
}

func TestInitNetwork(t *testing.T) {
	bootstrapID := NewKademliaID("2111111400000000000000000000000000000000")
	node := InitNode(bootstrapID)

	network := InitNetwork(node, 5)

	if network == nil {
		t.Errorf("Expected network to be initialized, but got nil.")
	}
}

func TestDecodeMessageHeaderWithEmptyData(t *testing.T) {
	_, err := DecodeMessageHeader([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, but got none")
	}
}

func TestDeserializeMessageWithEmptyData(t *testing.T) {
	_, err := DeserializeMessage([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, but got none")
	}
}

func TestEncodeDecodeMessageHeader(t *testing.T) {
	header := MessageHeader{
		HeaderLength: uint32(binary.Size(MessageHeader{})),
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    1,
		Type:         3,
		MessageID:    *NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
		SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
		ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
		BodyLength:   120,
		SenderIP:     StringTo15ByteArray(GetLocalIP()),
		ReceiverIP:   StringTo15ByteArray("192.168.0.2"),
	}

	encodedData, err := EncodeMessageHeader(header)
	if err != nil {
		t.Fatalf("Failed to encode MessageHeader: %v", err)
	}

	decodedHeader, err := DecodeMessageHeader(encodedData)
	if err != nil {
		t.Fatalf("Failed to decode MessageHeader: %v", err)
	}

	if header != decodedHeader {
		t.Error("Encoded and decoded headers do not match")
	}
}

func TestSerializeDeserializeMessage(t *testing.T) {
	data := []byte("hej")

	messageData := MessageData{
		Header: MessageHeader{
			HeaderLength: uint32(binary.Size(MessageHeader{})),
			HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
			Direction:    1,
			Type:         3,
			MessageID:    *NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
			SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
			ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
			BodyLength:   uint32(len(data)),
			SenderIP:     StringTo15ByteArray(GetLocalIP()),
			ReceiverIP:   StringTo15ByteArray("192.168.0.2"),
		},
		Data: data,
	}

	serialized, err := SerializeMessage(messageData)
	if err != nil {
		t.Fatalf("Failed to serialize message: %v", err)
	}

	deserialized, err := DeserializeMessage(serialized)
	if err != nil {
		t.Fatalf("Failed to deserialize message: %v", err)
	}

	if messageData.Header != deserialized.Header || !bytes.Equal(messageData.Data, deserialized.Data) {
		t.Errorf("Expected %v, got %v", messageData, deserialized)
	}
}
