package ipc

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessageType represents the type of IPC message
type MessageType int

const (
	MessageTypeRequest MessageType = iota
	MessageTypeResponse
	MessageTypeEvent
	MessageTypeError
)

// Message represents an IPC message
type Message struct {
	ID        string
	Type      MessageType
	From      string
	To        string
	Payload   interface{}
	Timestamp time.Time
	Error     error
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, from, to string, payload interface{}) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		From:      from,
		To:        to,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(from, to string, err error) *Message {
	return &Message{
		ID:        generateMessageID(),
		Type:      MessageTypeError,
		From:      from,
		To:        to,
		Error:     err,
		Timestamp: time.Now(),
	}
}

// Serialize serializes the message to JSON
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize deserializes a message from JSON
func Deserialize(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Request represents a request message
type Request struct {
	Method string
	Params map[string]interface{}
}

// Response represents a response message
type Response struct {
	Result interface{}
	Error  string
}

// Event represents an event message
type Event struct {
	Name    string
	Payload interface{}
}

var messageIDCounter uint64
var messageIDMu sync.Mutex

func generateMessageID() string {
	messageIDMu.Lock()
	defer messageIDMu.Unlock()
	messageIDCounter++
	return fmt.Sprintf("msg-%d", messageIDCounter)
}

