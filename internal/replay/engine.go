package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Event represents a recorded event
type Event struct {
	ID        string
	Type      string
	Timestamp time.Time
	Data      json.RawMessage
	Result    json.RawMessage
}

// ReplayEngine provides deterministic replay functionality
type ReplayEngine struct {
	events    []*Event
	current   int
	recording bool
	replaying bool
	mu        sync.RWMutex
}

// NewReplayEngine creates a new replay engine
func NewReplayEngine() *ReplayEngine {
	return &ReplayEngine{
		events: make([]*Event, 0),
	}
}

// StartRecording starts recording events
func (re *ReplayEngine) StartRecording() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.recording = true
	re.events = make([]*Event, 0)
	re.current = 0
}

// StopRecording stops recording events
func (re *ReplayEngine) StopRecording() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.recording = false
}

// RecordEvent records an event
func (re *ReplayEngine) RecordEvent(eventType string, data interface{}) (*Event, error) {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	if !re.recording {
		return nil, fmt.Errorf("not recording")
	}
	
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	event := &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      dataJSON,
	}
	
	re.events = append(re.events, event)
	return event, nil
}

// RecordResult records the result of an event
func (re *ReplayEngine) RecordResult(eventID string, result interface{}) error {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	if !re.recording {
		return fmt.Errorf("not recording")
	}
	
	for _, event := range re.events {
		if event.ID == eventID {
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}
			event.Result = resultJSON
			return nil
		}
	}
	
	return fmt.Errorf("event not found: %s", eventID)
}

// StartReplay starts replaying events
func (re *ReplayEngine) StartReplay() error {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	if len(re.events) == 0 {
		return fmt.Errorf("no events to replay")
	}
	
	re.replaying = true
	re.current = 0
	return nil
}

// StopReplay stops replaying
func (re *ReplayEngine) StopReplay() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.replaying = false
}

// NextEvent gets the next event to replay
func (re *ReplayEngine) NextEvent() (*Event, error) {
	re.mu.Lock()
	defer re.mu.Unlock()
	
	if !re.replaying {
		return nil, fmt.Errorf("not replaying")
	}
	
	if re.current >= len(re.events) {
		return nil, fmt.Errorf("no more events")
	}
	
	event := re.events[re.current]
	re.current++
	return event, nil
}

// Save saves events to a file
func (re *ReplayEngine) Save(filename string) error {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	data, err := json.MarshalIndent(re.events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}
	
	return os.WriteFile(filename, data, 0644)
}

// Load loads events from a file
func (re *ReplayEngine) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	var events []*Event
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}
	
	re.mu.Lock()
	defer re.mu.Unlock()
	re.events = events
	return nil
}

// GetEvents returns all recorded events
func (re *ReplayEngine) GetEvents() []*Event {
	re.mu.RLock()
	defer re.mu.RUnlock()
	
	result := make([]*Event, len(re.events))
	copy(result, re.events)
	return result
}

var eventIDCounter uint64
var eventIDMu sync.Mutex

func generateEventID() string {
	eventIDMu.Lock()
	defer eventIDMu.Unlock()
	eventIDCounter++
	return fmt.Sprintf("event-%d", eventIDCounter)
}

