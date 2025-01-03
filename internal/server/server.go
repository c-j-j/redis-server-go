package server

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"redis-go/internal/message"
	"redis-go/internal/parser"
	"strings"
)

func handleMessages(inMemoryDB map[string]string, writeChannel chan WriteRequest) {
	for nextMessage := range writeChannel {
		inMemoryDB[nextMessage.key] = nextMessage.value
	}
}

type WriteRequest struct {
	key   string
	value string
}

func StartServer() {
	if os.Getenv("VERBOSE") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	port := "6379"
	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server %v\n", err)
		os.Exit(-1)
	}
	defer listener.Close()

	writeChannel := make(chan WriteRequest)
	inMemoryDB := make(map[string]string)

	go handleMessages(inMemoryDB, writeChannel)

	fmt.Printf("Server started on port %s\n", port)
	for true {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection %v\n", err)
			os.Exit(-1)
		}

		go handleRequest(connection, writeChannel, inMemoryDB)
	}
}

func handleRequest(connection net.Conn, writeChannel chan WriteRequest, inMemoryDB map[string]string) {
	defer connection.Close()
	slog.Debug("Received new connection")

	// scanner := bufio.NewScanner(connection)
	buffer := make([]byte, 4096)

	for {
		n, err := connection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				slog.Debug("Client disconnected")
				break
			}
			slog.Error(fmt.Sprintf("Error reading request: %v", err))
			connection.Write([]byte("+Error reading request"))
			break
		}

		input := buffer[:n]

		message, err := parser.ReadMessage(string(input))

		if err != nil {
			slog.Error(fmt.Sprintf("Error parsing request: %v\n", err))
			connection.Write([]byte("+Error parsing request"))
			continue
		} else if message.Completed {
			response := handleResponse(message, writeChannel, inMemoryDB)
			connection.Write([]byte(response))
		}
	}
}

func handleResponse(message *message.Message, writeChannel chan WriteRequest, inMemoryDB map[string]string) string {
	tokens := message.Tokens
	messageType := tokens[0]

	switch strings.ToUpper(messageType) {
	case "PING":
		return "+PONG\r\n"
	case "ECHO":
		if len(tokens) != 2 {
			return "+Expected 1 argument provided to ECHO\r\n"
		}
		echoed := tokens[1]
		return fmt.Sprintf("$%d\r\n%s\r\n", len(echoed), echoed)
	case "SET":
		if len(tokens) != 3 {
			return "+Expected 2 arguments provided to SET\r\n"
		}
		key := tokens[1]
		value := tokens[2]
		writeChannel <- WriteRequest{key, value}
		return "+OK\r\n"
	case "GET":
		if len(tokens) != 2 {
			return "+Expected 1 argument provided to GET\r\n"
		}
		key := tokens[1]
		value, ok := inMemoryDB[key]
		if !ok {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	default:
		return "+UNSUPPORTED\r\n"
	}
}
