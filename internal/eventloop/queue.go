package eventloop

import (
	"container/heap"
	"sync"
)

// EventQueue is a priority queue for events
type EventQueue struct {
	events []*Event
	mu     sync.Mutex
	idGen  uint64
}

// NewEventQueue creates a new event queue
func NewEventQueue() *EventQueue {
	eq := &EventQueue{
		events: make([]*Event, 0),
	}
	heap.Init(eq)
	return eq
}

// Enqueue adds an event to the queue
func (eq *EventQueue) Enqueue(event *Event) {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	
	event.ID = eq.idGen
	eq.idGen++
	heap.Push(eq, event)
}

// Dequeue removes and returns the highest priority event
func (eq *EventQueue) Dequeue() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	
	if eq.Len() == 0 {
		return nil
	}
	
	return heap.Pop(eq).(*Event)
}

// Peek returns the highest priority event without removing it
func (eq *EventQueue) Peek() *Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	
	if eq.Len() == 0 {
		return nil
	}
	
	return eq.events[0]
}

// Len returns the number of events in the queue
func (eq *EventQueue) Len() int {
	return len(eq.events)
}

// Empty returns true if the queue is empty
func (eq *EventQueue) Empty() bool {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return len(eq.events) == 0
}

// Clear removes all events from the queue
func (eq *EventQueue) Clear() {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	eq.events = make([]*Event, 0)
	heap.Init(eq)
}

// Heap interface implementation
func (eq *EventQueue) Less(i, j int) bool {
	// Higher priority events come first
	if eq.events[i].Priority != eq.events[j].Priority {
		return eq.events[i].Priority > eq.events[j].Priority
	}
	// Earlier events come first if same priority
	return eq.events[i].Timestamp.Before(eq.events[j].Timestamp)
}

func (eq *EventQueue) Swap(i, j int) {
	eq.events[i], eq.events[j] = eq.events[j], eq.events[i]
}

func (eq *EventQueue) Push(x interface{}) {
	eq.events = append(eq.events, x.(*Event))
}

func (eq *EventQueue) Pop() interface{} {
	old := eq.events
	n := len(old)
	event := old[n-1]
	eq.events = old[0 : n-1]
	return event
}

// BackpressureThreshold is the maximum number of events before backpressure kicks in
const BackpressureThreshold = 10000

// IsOverloaded checks if the queue is overloaded
func (eq *EventQueue) IsOverloaded() bool {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return len(eq.events) > BackpressureThreshold
}

// Size returns the current queue size
func (eq *EventQueue) Size() int {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return len(eq.events)
}

