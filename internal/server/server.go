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
	"sync"
)

type WriteRequest struct {
	key   string
	value string
}

type InMemoryDB struct {
	data sync.Map
}

func newInMemoryDB() *InMemoryDB {
	db := InMemoryDB{
		data: sync.Map{},
	}
	return &db
}

func (db *InMemoryDB) writeValue(key string, value string) {
	db.data.Store(key, value)
}

func (db *InMemoryDB) getValue(key string) (string, bool) {
	value, ok := db.data.Load(key)
	return fmt.Sprintf("%v", value), ok
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

	inMemoryDB := newInMemoryDB()

	fmt.Printf("Server started on port %s\n", port)
	for {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection %v\n", err)
			os.Exit(-1)
		}

		go handleRequest(connection, inMemoryDB)
	}
}

func handleRequest(connection net.Conn, inMemoryDB *InMemoryDB) {
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
			connection.Write([]byte("-Error reading request"))
			break
		}

		input := buffer[:n]

		message, err := parser.ReadMessage(string(input))

		if err != nil {
			slog.Error(fmt.Sprintf("Error parsing request: %v\n", err))
			connection.Write([]byte("-Error parsing request"))
			continue
		} else if message.Completed {
			response := handleResponse(message, inMemoryDB)
			connection.Write([]byte(response))
		}
	}
}

func handleResponse(message *message.Message, inMemoryDB *InMemoryDB) string {
	tokens := message.Tokens
	messageType := tokens[0]

	switch strings.ToUpper(messageType) {
	case "PING":
		return "+PONG\r\n"
	case "ECHO":
		if len(tokens) != 2 {
			return "-Wrong number of arguments for 'ECHO' command\r\n"
		}
		echoed := tokens[1]
		return fmt.Sprintf("$%d\r\n%s\r\n", len(echoed), echoed)
	case "SET":
		if len(tokens) != 3 {
			return "-Wrong number of arguments for 'SET' command\r\n"
		}
		key := tokens[1]
		value := tokens[2]
		inMemoryDB.writeValue(key, value)
		return "+OK\r\n"
	case "GET":
		if len(tokens) != 2 {
			return "-Wrong number of arguments for 'GET' command\r\n"
		}
		key := tokens[1]
		value, ok := inMemoryDB.getValue(key)
		if !ok {
			return "$-1\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	default:
		return "-Unsupported command\r\n"
	}
}
