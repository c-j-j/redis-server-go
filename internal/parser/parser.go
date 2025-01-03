package parser

import (
	"errors"
	"fmt"
	"log/slog"
	"redis-go/internal/message"
	"strconv"
	"strings"
)

type Parser struct {
	nextArrayLength  *int
	currentMessage   message.Message
	nextStringLength *int
}

func (p *Parser) Reset() {
	*p = *NewParser()
}

func NewParser() *Parser {
	parser := Parser{
		nextArrayLength:  nil,
		nextStringLength: nil,
		currentMessage:   *message.NewMessage(),
	}
	return &parser
}

func ReadMessage(message string) (*message.Message, error) {
	lines := strings.Split(message, "\r\n")
	parser := NewParser()

	slog.Debug("HERE:")
	for _, line := range lines {
		message = strings.TrimSpace(line)
		slog.Debug("MESSAGE:")
		slog.Debug(message)

		if message == "" {
			parser.currentMessage.Completed = true
		} else if message[0] == '*' {
			rest := message[1:]
			length, err := strconv.Atoi(rest)

			if err != nil {
				fmt.Printf("Error converting '%s' to int, because of error %v", rest, err)
				return nil, errors.New("Invalid message")
			}
			parser.nextArrayLength = &length
		} else if message[0] == '$' {
			rest := message[1:]
			length, err := strconv.Atoi(rest)

			if err != nil {
				fmt.Printf("Error converting '%s' to int, because of error %v", rest, err)
				return nil, errors.New("Invalid message")
			}
			parser.nextStringLength = &length
		} else if parser.nextStringLength != nil {
			if *parser.nextStringLength != len(message) {
				slog.Error(fmt.Sprintf("Received string %s has unexpected length. Expected %d", message, *parser.nextStringLength))
				return nil, errors.New("Invalid message")
			}
			parser.currentMessage.Tokens = append(parser.currentMessage.Tokens, message)

			if len(parser.currentMessage.Tokens) == *parser.nextArrayLength {
				parser.currentMessage.Completed = true
			}
		}
	}

	return &parser.currentMessage, nil
}
