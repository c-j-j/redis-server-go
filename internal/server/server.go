package server

import (
	"fmt"
	"log/slog"
	"net"
	"os"
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

	for {
		nextMessage, err := parser.ParseInput(connection)
		if err != nil {
			slog.Error(fmt.Sprintf("Error reading request: %v", err))
			connection.Write([]byte("-Error reading request"))
			break
		}

		response := handleResponse(nextMessage, inMemoryDB)
		slog.Debug("Sending response: " + response.ResponseString())
		connection.Write([]byte(response.ResponseString()))
	}
}

func handleResponse(message parser.RespMessage, inMemoryDB *InMemoryDB) parser.RespMessage {
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
				return handleECHO(msg)
			case "GET":
				return handleGET(msg, inMemoryDB)
			case "SET":
				return handleSET(msg, inMemoryDB)
			default:
				slog.Debug("Invalid command sent to array", "msg", msg[0])
				return parser.NewError("Unexpected command sent in array")
			}
		}
	default:
		return parser.NewError("Unexpected command")
	}
}

func handleECHO(msg parser.RespArray) parser.RespMessage {
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

func handleSET(msg parser.RespArray, inMemoryDB *InMemoryDB) parser.RespMessage {
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
	inMemoryDB.writeValue(parsedKey, parsedValue)
	return parser.RespSimpleString("OK")
}

func handleGET(msg parser.RespArray, inMemoryDB *InMemoryDB) parser.RespMessage {
	if len(msg) < 2 {
		return parser.NewError("Wrong number of arguments for 'GET' command")
	}
	key := msg[1]
	if key, ok := key.(parser.RespBulkString); ok {
		value, ok := inMemoryDB.getValue(string(key))
		if !ok {
			return parser.NewBulkString(nil)
		}
		return parser.NewBulkString([]byte(value))
	} else {
		return parser.NewError("Expected string passed to ECHO")
	}
}
