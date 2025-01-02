package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"redis-go/internal/message"
	"redis-go/internal/parser"
	"strings"
)

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
	fmt.Printf("Server started on port %s\n", port)
	for true {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection %v\n", err)
			os.Exit(-1)
		}

		go handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	defer connection.Close()
	slog.Debug("Received new connection")

	scanner := bufio.NewScanner(connection)
	parser := parser.NewParser()

	for scanner.Scan() {
		input := scanner.Text()
		slog.Debug(input)

		message, err := parser.ReadNext(input)

		if err != nil {
			slog.Error(fmt.Sprintf("Error parsing request: %v\n", err))
			connection.Write([]byte("+Error parsing request"))
			parser.Reset()
			continue
		} else if message.Completed {
			response := generateResponse(message)
			parser.Reset()
			connection.Write([]byte(response))
		}
	}
}

func generateResponse(message *message.Message) string {
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
	default:
		return "+UNSUPPORTED\r\n"
	}
}
