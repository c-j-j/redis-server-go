package parser

import (
	"errors"
	"fmt"
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
	p = NewParser()
}

func NewParser() *Parser {
	parser := Parser{
		nextArrayLength:  nil,
		nextStringLength: nil,
		currentMessage:   *message.NewMessage(),
	}
	return &parser
}

func (p *Parser) ReadNext(message string) (*message.Message, error) {
	message = strings.TrimSpace(message)
	if message[0] == '*' {
		rest := message[1:]
		length, err := strconv.Atoi(rest)

		if err != nil {
			fmt.Printf("Error converting '%s' to int, because of error %v", rest, err)
			return nil, errors.New("Invalid message")
		}
		p.nextArrayLength = &length
	} else if message[0] == '$' {
		rest := message[1:]
		length, err := strconv.Atoi(rest)

		if err != nil {
			fmt.Printf("Error converting '%s' to int, because of error %v", rest, err)
			return nil, errors.New("Invalid message")
		}
		p.nextStringLength = &length
	} else if p.nextStringLength != nil {
		if *p.nextStringLength != len(message) {
			fmt.Printf("Received string %s has unexpected length. Expected %d", message, *p.nextStringLength)
			return nil, errors.New("Invalid message")
		}
		p.currentMessage.Tokens = append(p.currentMessage.Tokens, message)

		if len(p.currentMessage.Tokens) == *p.nextArrayLength {
			p.currentMessage.Completed = true
		}
	}

	return &p.currentMessage, nil
}
