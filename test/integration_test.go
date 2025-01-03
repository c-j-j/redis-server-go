package test

import (
	"fmt"
	"net"
	"redis-go/internal/server"
	"testing"
	"time"
)

var serverStarted = false

func startRedisServer() {
	if !serverStarted {
		go func() {
			server.StartServer()
			// Give the server time to start
			time.Sleep(1 * time.Second)
		}()
	}
	serverStarted = true
}

func TestRedisEchoCommand(t *testing.T) {
	startRedisServer()

	conn, err := net.Dial("tcp", "localhost:6379") // Adjust the port if your server uses a different one
	if err != nil {
		t.Fatalf("Failed to connect to Redis server: %v", err)
	}
	defer conn.Close()

	// Send ECHO command to the server
	command := "*2\r\n$4\r\nECHO\r\n$3\r\nHEY\r\n"
	fmt.Println("a")
	_, err = conn.Write([]byte(command))
	if err != nil {
		t.Fatalf("Failed to send command to Redis server: %v", err)
	}

	fmt.Println("hey")
	// Read the entire response from the server
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	response := string(buffer[:n])
	fmt.Println("bye")
	if err != nil {
		t.Fatalf("Failed to read response from Redis server: %v", err)
	}

	expectedResponse := "$3\r\nHEY\r\n"
	if string(response) != expectedResponse {
		t.Errorf("Unexpected response: got %q, want %q", response, expectedResponse)
	}
}

func TestRedisSetGetCommand(t *testing.T) {
	startRedisServer()

	conn, err := net.Dial("tcp", "localhost:6379") // Adjust the port if your server uses a different one
	if err != nil {
		t.Fatalf("Failed to connect to Redis server: %v", err)
	}
	defer conn.Close()

	// Send SET command to the server
	setCommand := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	_, err = conn.Write([]byte(setCommand))
	if err != nil {
		t.Fatalf("Failed to send SET command to Redis server: %v", err)
	}

	// Read response for SET command
	setBuffer := make([]byte, 4096)
	n, err := conn.Read(setBuffer)
	setResponse := string(setBuffer[:n])
	if err != nil {
		t.Fatalf("Failed to read SET response from Redis server: %v", err)
	}

	if setResponse != "+OK\r\n" {
		t.Errorf("Unexpected SET response: got %q, want %q", setResponse, "+OK\r\n")
	}

	// Send GET command to the server
	getCommand := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	_, err = conn.Write([]byte(getCommand))
	if err != nil {
		t.Fatalf("Failed to send GET command to Redis server: %v", err)
	}

	// Read response for GET command
	getBuffer := make([]byte, 4096)
	n, err = conn.Read(getBuffer)
	getResponse := string(getBuffer[:n])
	if err != nil {
		t.Fatalf("Failed to read GET response from Redis server: %v", err)
	}

	expectedGetResponse := "$5\r\nvalue\r\n"
	if getResponse != expectedGetResponse {
		t.Errorf("Unexpected GET response: got %q, want %q", getResponse, expectedGetResponse)
	}
}
