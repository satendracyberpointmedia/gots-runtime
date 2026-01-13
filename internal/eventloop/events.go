package eventloop

import (
	"time"
)

// EventType represents the type of event
type EventType int

const (
	EventIO EventType = iota
	EventTimer
	EventImmediate
	EventNextTick
)

// Event represents an event in the event loop
type Event struct {
	Type      EventType
	Handler   func() error
	Priority  int
	Timestamp time.Time
	ID        uint64
}

// NewEvent creates a new event
func NewEvent(eventType EventType, handler func() error, priority int) *Event {
	return &Event{
		Type:      eventType,
		Handler:   handler,
		Priority:  priority,
		Timestamp: time.Now(),
	}
}

// Execute executes the event handler
func (e *Event) Execute() error {
	if e.Handler == nil {
		return nil
	}
	return e.Handler()
}

// EventCallback is a function that handles an event
type EventCallback func() error

// TimerEvent represents a timer event
type TimerEvent struct {
	*Event
	Duration time.Duration
	Repeat   bool
}

// NewTimerEvent creates a new timer event
func NewTimerEvent(duration time.Duration, repeat bool, handler func() error) *TimerEvent {
	return &TimerEvent{
		Event:    NewEvent(EventTimer, handler, 0),
		Duration: duration,
		Repeat:   repeat,
	}
}

