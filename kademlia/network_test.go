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

func TestEncodeMessageHeaderWithBoundaryValues(t *testing.T) {
	header := MessageHeader{
		HeaderLength: 0xFFFFFFFF, // Maximum possible uint32 value
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    1,
		Type:         3,
		MessageID:    *NewKademliaID("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
		SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
		ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
		BodyLength:   0xFFFFFFFF, // Maximum possible uint32 value
		SenderIP:     StringTo15ByteArray(GetLocalIP()),
		ReceiverIP:   StringTo15ByteArray("192.168.0.2"),
	}

	encodedData, err := EncodeMessageHeader(header)
	if err != nil {
		t.Fatalf("Failed to encode boundary MessageHeader: %v", err)
	}

	decodedHeader, err := DecodeMessageHeader(encodedData)
	if err != nil {
		t.Fatalf("Failed to decode boundary MessageHeader: %v", err)
	}

	if header != decodedHeader {
		t.Error("Encoded and decoded boundary headers do not match")
	}
}

func TestDecodeMessageHeaderWithPartialData(t *testing.T) {
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

	partialData := encodedData[:len(encodedData)/2]

	_, err = DecodeMessageHeader(partialData)
	if err == nil {
		t.Error("Expected error when decoding partial data, but got none")
	}
}

func TestEncodeMessageHeaderWithInvalidData(t *testing.T) {
	header := MessageHeader{
		HeaderLength: 0xFFFFFFFF, // Exceeding value for testing
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    MessageDirection(255), // Invalid direction
		Type:         MessageType(255),      // Invalid type
		MessageID:    *NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
		SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
		ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
		BodyLength:   0xFFFFFFFF,
		SenderIP:     [15]byte{}, // Empty or invalid IP
		ReceiverIP:   [15]byte{}, // Empty or invalid IP
	}

	_, err := EncodeMessageHeader(header)
	if err == nil {
		t.Fatal("Expected error with invalid MessageHeader values, but got none")
	}
}

func TestEncodeMessageHeaderWithNilInput(t *testing.T) {
	var header MessageHeader
	encodedData, err := EncodeMessageHeader(header)
	if err != nil {
		t.Fatalf("Failed to encode empty MessageHeader: %v", err)
	}

	decodedHeader, err := DecodeMessageHeader(encodedData)
	if err != nil {
		t.Fatalf("Failed to decode empty MessageHeader: %v", err)
	}

	if header != decodedHeader {
		t.Error("Encoded and decoded headers for empty MessageHeader do not match")
	}
}

func TestDecodeMessageHeaderError(t *testing.T) {
	// Case 1: Pass in an empty byte array (invalid size)
	data := []byte{}
	_, err := DecodeMessageHeader(data)
	if err == nil {
		t.Error("Expected error with empty byte array, but got none")
	}

	// Case 2: Pass in a malformed byte array
	malformedData := []byte{0x01, 0x02, 0x03} // Too short to be valid
	_, err = DecodeMessageHeader(malformedData)
	if err == nil {
		t.Error("Expected error with malformed byte array, but got none")
	}
}

func TestByte15ArrayToString(t *testing.T) {
	// Test case 1: Normal conversion with no null bytes
	input := [15]byte{'H', 'e', 'l', 'l', 'o'}
	expected := "Hello"
	result := Byte15ArrayToString(input)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// Test case 2: Array filled with null bytes
	input = [15]byte{}
	expected = ""
	result = Byte15ArrayToString(input)
	if result != expected {
		t.Errorf("Expected an empty string, but got %s", result)
	}

	// Test case 3: Array with some characters and trailing null bytes
	input = [15]byte{'H', 'i', '\x00', '\x00', '\x00', '\x00'}
	expected = "Hi"
	result = Byte15ArrayToString(input)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// Test case 4: Full array with characters and no null bytes
	input = [15]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'A', 'B', 'C', 'D', 'E'}
	expected = "1234567890ABCDE"
	result = Byte15ArrayToString(input)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// Test case 5: A 15-byte array where "Test" is followed by null characters.
	var testBuffer [15]byte
	copy(testBuffer[:], "Test")
	result = Byte15ArrayToString(testBuffer)
	expected = "Test"
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestSerializeMessageError(t *testing.T) {
	// Create a MessageData with an invalid header to cause EncodeMessageHeader to fail
	messageData := MessageData{
		Header: MessageHeader{
			HeaderLength: uint32(binary.Size(MessageHeader{})),
			HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
			Direction:    MessageDirection(255), // Invalid direction to cause an error
			Type:         MessageType(255),      // Invalid type to cause an error
			MessageID:    *NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
			SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
			ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
			BodyLength:   120,
			SenderIP:     [15]byte{0},
			ReceiverIP:   [15]byte{0},
		},
		Data: []byte("test data"),
	}

	// Call SerializeMessage with the invalid header
	_, err := SerializeMessage(messageData)
	if err == nil {
		t.Error("Expected error with invalid header in SerializeMessage, but got none")
	}
}

func TestEncodeMessageHeaderInvalidTypeAndDirection(t *testing.T) {
	header := MessageHeader{
		HeaderLength: uint32(binary.Size(MessageHeader{})),
		HeaderTag:    [4]byte{'K', 'A', 'D', 'M'},
		Direction:    0,
		Type:         MAX_MESSAGE_TYPE,
		MessageID:    *NewKademliaID("FFFFFFFF00000000000000000000000000000000"),
		SenderID:     *NewKademliaID("1111111100000000000000000000000000000000"),
		ReceiverID:   *NewKademliaID("1111111200000000000000000000000000000000"),
		BodyLength:   120,
		SenderIP:     StringTo15ByteArray(GetLocalIP()),
		ReceiverIP:   StringTo15ByteArray("192.168.0.2"),
	}

	// Invalid Message Type
	header.Type = MAX_MESSAGE_TYPE
	_, err := EncodeMessageHeader(header)
	if err == nil || err.Error() != "invalid MessageType: 4" {
		t.Errorf("Expected invalid MessageType error, got %v", err)
	}

	// Reset header for the next test
	header.Type = 0

	// Invalid Message Direction
	header.Direction = MAX_MESSAGE_DIRECTION
	_, err = EncodeMessageHeader(header)
	if err == nil || err.Error() != "invalid MessageDirection: 2" {
		t.Errorf("Expected invalid MessageDirection error, got %v", err)
	}

}
