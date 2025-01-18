package server

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"redis-go/internal/parser"
	"redis-go/internal/storage"
	"strconv"
	"strings"
)

type Server struct {
	db *storage.InMemoryDB
}

func StartServer() {
	if os.Getenv("VERBOSE") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	portArg := flag.Int("port", 6379, "Port to listen on")
	flag.Parse()

	port := strconv.Itoa(*portArg)

	fmt.Println("port", port)
	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server %v\n", err)
		os.Exit(-1)
	}
	defer listener.Close()
	server := initServer()

	fmt.Printf("Server started on port %s\n", port)
	for {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection %v\n", err)
			os.Exit(-1)
		}

		go server.handleRequest(connection)
	}
}

func initServer() Server {
	inMemoryDB := storage.NewInMemoryDB()
	return Server{
		db: inMemoryDB,
	}
}

func (s *Server) handleRequest(connection net.Conn) {
	defer connection.Close()
	slog.Debug("Received new connection")

	for {
		nextMessage, err := parser.ParseInput(connection)
		if err != nil {
			slog.Error(fmt.Sprintf("Error reading request: %v", err))
			connection.Write([]byte("-Error reading request"))
			break
		}

		response := s.handleResponse(nextMessage)
		slog.Debug("Sending response: " + response.ResponseString())
		connection.Write([]byte(response.ResponseString()))
	}
}

func (s *Server) handleResponse(message parser.RespMessage) parser.RespMessage {
	slog.Debug("Handling request")
	slog.Debug(parser.PrintRespMessage(message))
	switch msg := message.(type) {
	case parser.RespSimpleString:
		return parser.NewSimpleString("OK")
	case parser.RespArray:
		command := msg[0]
		if command, ok := command.(parser.RespBulkString); !ok {
			return parser.NewError("Wrong value sent in array command")
		} else {
			switch strings.ToUpper(string(command)) {
			case "PING":
				return parser.NewSimpleString("PONG")
			case "ECHO":
				return s.handleECHO(msg)
			case "GET":
				return s.handleGET(msg)
			case "SET":
				return s.handleSET(msg)
			case "INFO":
				return s.handleINFO(msg)
			default:
				slog.Debug("Invalid command sent to array", "msg", msg[0])
				return parser.NewError("Unexpected command sent in array")
			}
		}
	default:
		return parser.NewError("Unexpected command")
	}
}

func (s *Server) handleECHO(msg parser.RespArray) parser.RespMessage {
	if len(msg) != 2 {
		return parser.NewError("Wrong number of arguments for 'ECHO' command")
	}
	echoed := msg[1]
	if echoed, ok := echoed.(parser.RespBulkString); ok {
		return parser.NewBulkString(echoed)
	} else {
		return parser.NewError("Expected string passed to ECHO")
	}
}

func (s *Server) handleSET(msg parser.RespArray) parser.RespMessage {
	if len(msg) < 3 {
		return parser.NewError("Wrong number of arguments for 'SET' command")
	}
	key := msg[1]
	value := msg[2]
	parsedKey := ""
	parsedValue := ""
	if key, ok := key.(parser.RespBulkString); ok {
		parsedKey = string(key)
	} else {
		return parser.NewError("Expected string as key passed to SET")
	}

	if value, ok := value.(parser.RespBulkString); ok {
		parsedValue = string(value)
	} else {
		return parser.NewError("Expected string as value passed to SET")
	}
	s.db.WriteValue(parsedKey, parsedValue)
	return parser.RespSimpleString("OK")
}

func (s *Server) handleINFO(msg parser.RespArray) parser.RespMessage {
	// TODO: continue from here
	return parser.NewError("INFO not supported yet")
}

func (s *Server) handleGET(msg parser.RespArray) parser.RespMessage {
	if len(msg) < 2 {
		return parser.NewError("Wrong number of arguments for 'GET' command")
	}
	key := msg[1]
	if key, ok := key.(parser.RespBulkString); ok {
		value, ok := s.db.GetValue(string(key))
		if !ok {
			return parser.NewBulkString(nil)
		}
		return parser.NewBulkString([]byte(value))
	} else {
		return parser.NewError("Expected string passed to ECHO")
	}
}
