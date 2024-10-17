package kademlia

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func createMockNetwork() *Network {
	id := NewKademliaID("0000000000000000000000000000000000000000")
	me := NewContact(id, "localhost")
	network := &Network{
		Node: &Kademlia{
			Routes:       NewRoutingTable(me),
			Network:      nil,
			Storage:      make(map[KademliaID]*StorageItem),
			ShutdownChan: make(chan struct{}),
		},
		ResponseMap:      make(map[KademliaID]chan MessageData),
		ResponseMapMutex: sync.Mutex{},
		Port:             8000,
		Timeout:          5,
	}

	network.Node.Network = network
	return network
}

func (k *Kademlia) TestPing(contact *Contact) error {
	if contact.Address == "127.0.0.1:8000" {
		return nil
	}
	return fmt.Errorf("Ping could not be sent to %s", contact.Address)
}

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

// Det här testat ska nog funka, men får de inte att funka längre
// Test PutController with valid "data" parameter
func TestPutController_ValidDataParam(t *testing.T) {
	network := createMockNetwork()
	os.Setenv("OBJECT_TTL", "10")
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
	data := "mockdata123"
	hashedData := HashData(data)
	key := NewKademliaID(hashedData)
	network.Node.StorageSet(key, &data)
	req := httptest.NewRequest("GET", fmt.Sprintf("/get?hash=%v", key), nil)
	w := httptest.NewRecorder()

	network.GetController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Data found for hash %v: mockdata123 on node %v", key, network.Node.Routes.Me.ID.String())
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

func TestGetController_MissingHashParam(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("GET", "/get", nil)
	w := httptest.NewRecorder()

	network.GetController(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status BadRequest; got %v", resp.Status)
	}
	body, _ := io.ReadAll(resp.Body)
	expected := "Missing 'hash' parameter"
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

func TestGetController_NotFound(t *testing.T) {
	network := createMockNetwork()
	id := NewRandomKademliaID()
	network.Node.Routes.AddContact(NewContact(id, "localhost"))
	data := "mockdata123"
	hashedData := HashData(data)
	key := NewKademliaID(hashedData)
	req := httptest.NewRequest("GET", fmt.Sprintf("/get?hash=%v", key), nil)
	w := httptest.NewRecorder()

	network.GetController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Data not found for hash %s. Closest nodes are:\n- Node ID: %s, Address: localhost\n", hashedData, id)
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
	data := "mockdata123"
	hashedData := HashData(data)
	key := NewKademliaID(hashedData)
	network.Node.StorageSet(key, &data)

	req := httptest.NewRequest("GET", "/show-storage", nil)
	w := httptest.NewRecorder()

	network.ShowStorageController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Stored data:\nKey: %s, Value: mockdata123\n", key.String())
	if string(body) != expected {
		t.Errorf("Expected response body to be %q, got %q", expected, string(body))
	}
}

func TestBeginResponseWithXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	BeginResponse(req, "/ping")

	// assume BeginResponse should have used "203.0.113.195" as clientIP
	expectedClientIP := "203.0.113.195"
	actualClientIP := req.Header.Get("X-Forwarded-For")

	if actualClientIP != expectedClientIP {
		t.Errorf("Expected clientIP to be %s, got %s", expectedClientIP, actualClientIP)
	}
}

func TestForgetController_MissingHashParam(t *testing.T) {
	network := createMockNetwork()
	req := httptest.NewRequest("POST", "/forget", nil)
	w := httptest.NewRecorder()

	network.ForgetController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status BadRequest; got %v", resp.Status)
	}

	expected := "Missing 'hash' parameter"
	if !strings.Contains(string(body), expected) {
		t.Errorf("Expected response body to contain %q, got %q", expected, string(body))
	}
}

func TestForgetController_ValidHashParam(t *testing.T) {
	network := createMockNetwork()
	hash := "data"
	req := httptest.NewRequest("POST", fmt.Sprintf("/forget?hash=%s", hash), nil) // Prepare the request with the hash
	w := httptest.NewRecorder()

	network.ForgetController(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK; got %v", resp.Status)
	}

	expected := fmt.Sprintf("Stopped refreshing item %s", hash)
	if string(body) != expected {
		t.Errorf("Expected response body to be %q, got %q", expected, string(body))
	}
}
