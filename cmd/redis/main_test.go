package main

import (
	"bufio"
	"net"
	"testing"
	"time"
)

func startRedisServer() {
	// Replace this function with your actual Redis server startup logic.
	// For example, you might call your server's `main` function in a goroutine.
	go func() {
		main() // Assuming your Redis server has a main function
	}()
	// Give the server time to start
	time.Sleep(1 * time.Second)
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
	_, err = conn.Write([]byte(command))
	if err != nil {
		t.Fatalf("Failed to send command to Redis server: %v", err)
	}

	// Read response from the server
	reader := bufio.NewReader(conn)
	// Verify the line length
	lineLength, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response from Redis server: %v", err)
	}

	expectedLineLength := "$3\r\n"
	if lineLength != expectedLineLength {
		t.Errorf("Unexpected response: got %q, want %q", lineLength, expectedLineLength)
	}

	// Verify the data
	dataLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response from Redis server: %v", err)
	}

	// Verify the response
	expectedDataLine := "HEY\r\n"
	if dataLine != expectedDataLine {
		t.Errorf("Unexpected response: got %q, want %q", dataLine, expectedDataLine)
	}
}
