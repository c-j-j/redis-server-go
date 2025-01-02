package message

type Message struct {
	Tokens    []string
	Completed bool
}

func NewMessage() *Message {
	message := Message{
		Tokens:    make([]string, 0, 10),
		Completed: false,
	}
	return &message
}
