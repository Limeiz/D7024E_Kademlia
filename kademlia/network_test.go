package kademlia

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestInitNetwork(t *testing.T) {
	bootstrapID := NewKademliaID("2111111400000000000000000000000000000000")
	node := InitNode(bootstrapID)

	network := InitNetwork(node, 5)

	if network == nil {
		t.Errorf("Expected network to be initialized, but got nil.")
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
