package kademlia

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Helper function to create a mock network with the mock node
func createMockNetwork() *Network {
	id := NewKademliaID("0000000000000000000000000000000000000000")
	me := NewContact(id, "localhost")
	network := &Network{
		Node: &Kademlia{
			Routes:       NewRoutingTable(me), // Initialize RoutingTable if needed
			Network:      nil,                 // Will be set later
			Storage:      make(map[KademliaID]string),
			ShutdownChan: make(chan struct{}),
		},
		ResponseMap:      make(map[KademliaID]chan MessageData),
		ResponseMapMutex: sync.Mutex{},
		Port:             8000,
		Timeout:          5,
	}

	// Associate the network with the Kademlia node
	network.Node.Network = network
	return network
}

func (k *Kademlia) TestPing(contact *Contact) error {
	// Modify this logic based on your test case
	if contact.Address == "127.0.0.1:8000" {
		return nil // Simulate success for this address
	}
	return fmt.Errorf("Ping could not be sent to %s", contact.Address)
}

// Test DefaultController
func TestDefaultController(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	network.DefaultController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := "Try some of these commands"
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

// Test GetIDController
func TestGetIDController(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("GET", "/getid", nil)
	w := httptest.NewRecorder()

	network.GetID(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := "This node's id is: 0000000000000000000000000000000000000000"
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

// Test PingController with missing "to" parameter
func TestPingController_MissingToParam(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	network.PingController(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status BadRequest; got %v", resp.Status)
	}
}

// Test PutController with missing "data" parameter
func TestPutController_MissingDataParam(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("POST", "/put", nil)
	w := httptest.NewRecorder()

	network.PutController(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status BadRequest; got %v", resp.Status)
	}
}

// Test PutController with valid "data" parameter
func TestPutController_ValidDataParam(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("POST", "/put", strings.NewReader("data=sampledata"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	network.PutController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := "Error: Could not store data: No contacts found in the routing table."
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

// Test GetController with valid hash parameter
func TestGetController_ValidHash(t *testing.T) {
	network := createMockNetwork()
	hashedData := HashData("mockdata123")
	hash := NewKademliaID(hashedData)
	network.Node.Storage[*hash] = "mockdata123"
	req := httptest.NewRequest("GET", fmt.Sprintf("/get?hash=%v", hash), nil)
	w := httptest.NewRecorder()

	network.GetController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Data found for hash %v: mockdata123 on node %v", hash, network.Node.Routes.Me.ID.String())
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

func TestExitController(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("GET", "/exit", nil)
	w := httptest.NewRecorder()

	network.ExitController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := "Shutting down node..."
	if string(body) != expected {
		t.Errorf("Expected response body to be %q, got %q", expected, string(body))
	}
}

func TestShowRoutingTableController(t *testing.T) {
	network := createMockNetwork()

	req := httptest.NewRequest("GET", "/show-routing-table", nil)
	w := httptest.NewRecorder()

	network.ShowRoutingTableController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Routing Table:\n%s", network.Node.Routes.String())
	if string(body) != expected {
		t.Errorf("Expected response body to be %q, got %q", expected, string(body))
	}
}

func TestShowStorageController(t *testing.T) {
	network := createMockNetwork()
	hashedData := HashData("mockdata123")
	hash := NewKademliaID(hashedData)
	network.Node.Storage[*hash] = "mockdata123"

	req := httptest.NewRequest("GET", "/show-storage", nil)
	w := httptest.NewRecorder()

	network.ShowStorageController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Stored data:\nKey: %s, Value: mockdata123\n", hash.String())
	if string(body) != expected {
		t.Errorf("Expected response body to be %q, got %q", expected, string(body))
	}
}
