package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"redis-go/internal/message"
	"strconv"
	"strings"
)

type Parser struct {
	nextArrayLength  *int
	currentMessage   message.Message
	nextStringLength *int
}

type RespMessage interface {
	ResponseString() string
}
type RespSimpleString string
type RespError string
type RespInteger int
type RespBulkString []byte
type RespArray []RespMessage

func (message RespArray) ResponseString() string {
	panic("Responding with array not yet supported")
}
func (message RespError) ResponseString() string {
	return fmt.Sprintf("-%s\r\n", message)
}
func (message RespSimpleString) ResponseString() string {
	return fmt.Sprintf("+%s\r\n", message)
}
func (message RespInteger) ResponseString() string {
	panic("Responding with integer not yet supported")
}
func (message RespBulkString) ResponseString() string {
	if message == nil {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)
}

func NewBulkString(input []byte) RespBulkString {
	return input
}
func NewSimpleString(input string) RespSimpleString {
	return RespSimpleString(input)
}
func NewError(input string) RespSimpleString {
	return RespSimpleString(input)
}

func ParseInput(reader io.Reader) (RespMessage, error) {
	bufReader := bufio.NewReader(reader)
	nextByte, err := bufReader.ReadByte()

	if err != nil {
		return nil, err
	}

	switch nextByte {
	case '+':
		{
			line, _, err := bufReader.ReadLine()
			if err != nil {
				return nil, err
			}
			return RespSimpleString(line), nil
		}
	case ':':
		{
			line, _, err := bufReader.ReadLine()
			if err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(string(line))
			if err != nil {
				return nil, err
			}
			return RespInteger(n), nil
		}
	case '$':
		{
			line, _, err := bufReader.ReadLine()
			if err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(string(line))
			if err != nil {
				return nil, err
			}
			bulkString := make([]byte, n)
			io.ReadFull(bufReader, bulkString)
			// consume new line char
			bufReader.ReadLine()
			return RespBulkString(bulkString), nil
		}
	case '*':
		{
			line, _, err := bufReader.ReadLine()
			if err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(string(line))
			if err != nil {
				return nil, err
			}
			array := make(RespArray, n)

			for i := 0; i < n; i++ {
				next, err := ParseInput(bufReader)

				if err != nil {
					return nil, err
				}
				array[i] = next
			}
			return array, nil
		}
	default:
		{
			return nil, errors.New("Unexpected token " + string(nextByte))
		}
	}
}

func PrintRespMessage(msg RespMessage) string {
	switch v := msg.(type) {
	case RespSimpleString:
		return fmt.Sprintf("SimpleString: %s", string(v))
	case RespError:
		return fmt.Sprintf("Error: %s", string(v))
	case RespInteger:
		return fmt.Sprintf("Integer: %d", int(v))
	case RespBulkString:
		if v == nil {
			return "BulkString: (nil)"
		}
		return fmt.Sprintf("BulkString: %s", string(v))
	case RespArray:
		var elements []string
		for _, element := range v {
			elements = append(elements, PrintRespMessage(element))
		}
		return fmt.Sprintf("Array: [%s]", strings.Join(elements, ", "))
	default:
		return "Unknown type"
	}
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
