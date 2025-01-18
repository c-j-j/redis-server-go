package parser

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseRESP(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  RespMessage
		expectErr bool
	}{
		{
			name:      "Simple String",
			input:     "+OK\r\n",
			expected:  RespSimpleString("OK"),
			expectErr: false,
		},
		{
			name:      "Integer",
			input:     ":1000\r\n",
			expected:  RespInteger(1000),
			expectErr: false,
		},
		{
			name:      "Bulk String",
			input:     "$6\r\nfoobar\r\n",
			expected:  RespBulkString([]byte("foobar")),
			expectErr: false,
		},
		{
			name:      "Array",
			input:     "*3\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$3\r\nbaz\r\n",
			expected:  RespArray{RespBulkString([]byte("foo")), RespBulkString([]byte("bar")), RespBulkString([]byte("baz"))},
			expectErr: false,
		},
		// {
		// 	name:      "Null Array",
		// 	input:     "*-1\r\n",
		// 	expected:  nil,
		// 	expectErr: false,
		// },
		// {
		// 	name:      "Malformed Input",
		// 	input:     "+OK",
		// 	expected:  nil,
		// 	expectErr: true,
		// },
		{
			name:      "Invalid RESP Type",
			input:     "@Invalid\r\n",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			result, err := ParseInput(reader)

			if test.expectErr {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !compareMessages(result, test.expected) {
				t.Errorf("expected: %v, got: %v", test.expected, result)
			}
		})
	}
}

func compareMessages(a, b RespMessage) bool {
	switch a := a.(type) {
	case RespSimpleString:
		b, ok := b.(RespSimpleString)
		return ok && a == b
	case RespError:
		b, ok := b.(RespError)
		return ok && a == b
	case RespInteger:
		b, ok := b.(RespInteger)
		return ok && a == b
	case RespBulkString:
		b, ok := b.(RespBulkString)
		return ok && bytes.Equal(a, b)
	case RespArray:
		b, ok := b.(RespArray)
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !compareMessages(a[i], b[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// package parser
//
// import (
// 	"redis-go/internal/message"
// 	"testing"
// )
//
// func TestReadMessage_ValidSimpleString(t *testing.T) {
// 	input := "+OK\r\n"
// 	expected := &message.Message{
// 		Tokens:    []string{"OK"},
// 		Completed: true,
// 	}
//
// 	result, err := ReadMessage(input)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
//
// 	if !result.Completed {
// 		t.Fatalf("expected message to be completed")
// 	}
//
// 	if len(result.Tokens) != len(expected.Tokens) || result.Tokens[0] != expected.Tokens[0] {
// 		t.Fatalf("expected %v, got %v", expected.Tokens, result.Tokens)
// 	}
// }
//
// func TestReadMessage_ValidArrayMessage(t *testing.T) {
// 	input := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
// 	expected := &message.Message{
// 		Tokens:    []string{"foo", "bar"},
// 		Completed: true,
// 	}
//
// 	result, err := ReadMessage(input)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
//
// 	if !result.Completed {
// 		t.Fatalf("expected message to be completed")
// 	}
//
// 	if len(result.Tokens) != len(expected.Tokens) || result.Tokens[0] != expected.Tokens[0] || result.Tokens[1] != expected.Tokens[1] {
// 		t.Fatalf("expected %v, got %v", expected.Tokens, result.Tokens)
// 	}
// }
//
// func TestReadMessage_InvalidArrayLength(t *testing.T) {
// 	input := "*2\r\n$3\r\nfoo\r\n$3\r\n"
// 	_, err := ReadMessage(input)
//
// 	if err == nil {
// 		t.Fatalf("expected error for incomplete array message")
// 	}
// }
//
// func TestReadMessage_InvalidStringLength(t *testing.T) {
// 	input := "*1\r\n$3\r\nf\r\n"
// 	_, err := ReadMessage(input)
//
// 	if err == nil {
// 		t.Fatalf("expected error for invalid string length")
// 	}
// }
//
// func TestReadMessage_EmptyMessage(t *testing.T) {
// 	input := ""
// 	result, err := ReadMessage(input)
//
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
//
// 	if !result.Completed {
// 		t.Fatalf("expected message to be completed")
// 	}
//
// 	if len(result.Tokens) != 0 {
// 		t.Fatalf("expected no tokens, got %v", result.Tokens)
// 	}
// }
