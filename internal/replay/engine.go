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

// GetEventCount returns the number of recorded events
func (re *ReplayEngine) GetEventCount() int {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return len(re.events)
}

// GetCurrentEventIndex returns the current replay position
func (re *ReplayEngine) GetCurrentEventIndex() int {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.current
}

// IsReplaying checks if currently replaying
func (re *ReplayEngine) IsReplaying() bool {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.replaying
}

// IsRecording checks if currently recording
func (re *ReplayEngine) IsRecording() bool {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.recording
}

// Clear clears all recorded events
func (re *ReplayEngine) Clear() {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.events = make([]*Event, 0)
	re.current = 0
	re.recording = false
	re.replaying = false
}

// GetEventsByType returns all events of a specific type
func (re *ReplayEngine) GetEventsByType(eventType string) []*Event {
	re.mu.RLock()
	defer re.mu.RUnlock()

	result := make([]*Event, 0)
	for _, event := range re.events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}
	return result
}

// GetEventByID gets a specific event by ID
func (re *ReplayEngine) GetEventByID(eventID string) (*Event, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	for _, event := range re.events {
		if event.ID == eventID {
			return event, true
		}
	}
	return nil, false
}

// ReplayStats contains statistics about replay execution
type ReplayStats struct {
	TotalEvents     int
	ExecutedEvents  int
	FailedEvents    int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	StartTime       time.Time
	EndTime         time.Time
}

// ExecutionRecord tracks a single replay execution
type ExecutionRecord struct {
	EventID   string
	Status    string // "pending", "executing", "completed", "failed"
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Error     error
	Result    interface{}
}

// ReplaySession manages a replay session
type ReplaySession struct {
	*ReplayEngine
	records   map[string]*ExecutionRecord
	stats     *ReplayStats
	sessionID string
	mu        sync.RWMutex
}

// NewReplaySession creates a new replay session
func NewReplaySession(sessionID string) *ReplaySession {
	return &ReplaySession{
		ReplayEngine: NewReplayEngine(),
		records:      make(map[string]*ExecutionRecord),
		sessionID:    sessionID,
		stats: &ReplayStats{
			StartTime: time.Now(),
		},
	}
}

// RecordExecution records the execution of an event
func (rs *ReplaySession) RecordExecution(eventID string, status string, duration time.Duration, err error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	record := &ExecutionRecord{
		EventID:   eventID,
		Status:    status,
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Duration:  duration,
		Error:     err,
	}

	rs.records[eventID] = record

	if err == nil {
		rs.stats.ExecutedEvents++
	} else {
		rs.stats.FailedEvents++
	}
}

// GetSessionStats returns session statistics
func (rs *ReplaySession) GetSessionStats() *ReplayStats {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	stats := *rs.stats
	if stats.ExecutedEvents > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(stats.ExecutedEvents)
	}
	stats.TotalEvents = len(rs.ReplayEngine.events)
	return &stats
}

// GetExecutionRecord gets the execution record for an event
func (rs *ReplaySession) GetExecutionRecord(eventID string) (*ExecutionRecord, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	record, ok := rs.records[eventID]
	return record, ok
}

var eventIDCounter uint64
var eventIDMu sync.Mutex

func generateEventID() string {
	eventIDMu.Lock()
	defer eventIDMu.Unlock()
	eventIDCounter++
	return fmt.Sprintf("event-%d", eventIDCounter)
}
