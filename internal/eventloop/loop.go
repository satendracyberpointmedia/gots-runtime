package eventloop

import (
	"context"
	"sync"
	"time"
)

// Loop represents the event loop
type Loop struct {
	queue       *EventQueue
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
	timers      map[uint64]*TimerEvent
	timerMu     sync.Mutex
	nextTick    []EventCallback
	nextTickMu  sync.Mutex
}

// NewLoop creates a new event loop
func NewLoop(ctx context.Context) *Loop {
	loopCtx, cancel := context.WithCancel(ctx)
	return &Loop{
		queue:   NewEventQueue(),
		ctx:     loopCtx,
		cancel:  cancel,
		timers:  make(map[uint64]*TimerEvent),
		nextTick: make([]EventCallback, 0),
	}
}

// Start starts the event loop
func (l *Loop) Start() {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return
	}
	l.running = true
	l.mu.Unlock()

	l.wg.Add(1)
	go l.run()
}

// Stop stops the event loop
func (l *Loop) Stop() {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return
	}
	l.running = false
	l.mu.Unlock()

	l.cancel()
	l.wg.Wait()
}

// Enqueue adds an event to the queue
func (l *Loop) Enqueue(event *Event) error {
	if l.IsOverloaded() {
		return ErrQueueOverloaded
	}
	l.queue.Push(event)
	return nil
}

// SetTimeout schedules a function to run after a delay
func (l *Loop) SetTimeout(duration time.Duration, handler func() error) uint64 {
	timer := NewTimerEvent(duration, false, handler)
	l.timerMu.Lock()
	timerID := timer.ID
	l.timers[timerID] = timer
	l.timerMu.Unlock()

	// Schedule the timer
	go func() {
		select {
		case <-time.After(duration):
			l.Enqueue(timer.Event)
			l.timerMu.Lock()
			delete(l.timers, timerID)
			l.timerMu.Unlock()
		case <-l.ctx.Done():
			return
		}
	}()

	return timerID
}

// SetInterval schedules a function to run repeatedly
func (l *Loop) SetInterval(duration time.Duration, handler func() error) uint64 {
	timer := NewTimerEvent(duration, true, handler)
	l.timerMu.Lock()
	timerID := timer.ID
	l.timers[timerID] = timer
	l.timerMu.Unlock()

	// Schedule the repeating timer
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Enqueue(timer.Event)
			case <-l.ctx.Done():
				l.timerMu.Lock()
				delete(l.timers, timerID)
				l.timerMu.Unlock()
				return
			}
		}
	}()

	return timerID
}

// ClearTimeout clears a timeout
func (l *Loop) ClearTimeout(id uint64) {
	l.timerMu.Lock()
	defer l.timerMu.Unlock()
	delete(l.timers, id)
}

// ClearInterval clears an interval
func (l *Loop) ClearInterval(id uint64) {
	l.ClearTimeout(id)
}

// NextTick schedules a callback to run on the next tick
func (l *Loop) NextTick(callback EventCallback) {
	l.nextTickMu.Lock()
	defer l.nextTickMu.Unlock()
	l.nextTick = append(l.nextTick, callback)
}

// SetImmediate schedules a callback to run immediately
func (l *Loop) SetImmediate(callback EventCallback) {
	event := NewEvent(EventImmediate, callback, 10)
	l.Enqueue(event)
}

// IsOverloaded checks if the event queue is overloaded
func (l *Loop) IsOverloaded() bool {
	return l.queue.IsOverloaded()
}

// run is the main event loop
func (l *Loop) run() {
	defer l.wg.Done()

	for {
		select {
		case <-l.ctx.Done():
			return
		default:
			// Process nextTick callbacks first
			l.processNextTick()

		// Process events from queue
		event := l.queue.Dequeue()
		if event != nil {
			_ = event.Execute()
		} else {
			// No events, sleep briefly to avoid busy waiting
			time.Sleep(1 * time.Millisecond)
		}
		}
	}
}

// processNextTick processes all nextTick callbacks
func (l *Loop) processNextTick() {
	l.nextTickMu.Lock()
	callbacks := l.nextTick
	l.nextTick = make([]EventCallback, 0)
	l.nextTickMu.Unlock()

	for _, callback := range callbacks {
		_ = callback()
	}
}

// Errors
var (
	ErrQueueOverloaded = &EventLoopError{Message: "event queue is overloaded"}
)

// EventLoopError represents an event loop error
type EventLoopError struct {
	Message string
}

func (e *EventLoopError) Error() string {
	return e.Message
}

