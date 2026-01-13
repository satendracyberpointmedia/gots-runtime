package ipc

import (
	"context"
	"sync"
)

// Channel represents a communication channel
type Channel struct {
	name      string
	sendChan  chan *Message
	recvChan  chan *Message
	ctx       context.Context
	cancel    context.CancelFunc
	subscribers map[string]chan *Message
	mu        sync.RWMutex
}

// NewChannel creates a new channel
func NewChannel(name string, ctx context.Context) *Channel {
	channelCtx, cancel := context.WithCancel(ctx)
	return &Channel{
		name:        name,
		sendChan:    make(chan *Message, 100),
		recvChan:    make(chan *Message, 100),
		ctx:         channelCtx,
		cancel:      cancel,
		subscribers: make(map[string]chan *Message),
	}
}

// Send sends a message on the channel
func (c *Channel) Send(msg *Message) error {
	select {
	case c.sendChan <- msg:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
}

// Receive receives a message from the channel
func (c *Channel) Receive() (*Message, error) {
	select {
	case msg := <-c.recvChan:
		return msg, nil
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

// Subscribe subscribes to messages on this channel
func (c *Channel) Subscribe(subscriberID string) (<-chan *Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	subChan := make(chan *Message, 100)
	c.subscribers[subscriberID] = subChan
	return subChan, nil
}

// Unsubscribe unsubscribes from messages
func (c *Channel) Unsubscribe(subscriberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribers, subscriberID)
}

// Start starts the channel message routing
func (c *Channel) Start() {
	go c.route()
}

// Stop stops the channel
func (c *Channel) Stop() {
	c.cancel()
	close(c.sendChan)
	close(c.recvChan)
	
	c.mu.Lock()
	for _, subChan := range c.subscribers {
		close(subChan)
	}
	c.mu.Unlock()
}

// route routes messages to subscribers
func (c *Channel) route() {
	for {
		select {
		case msg, ok := <-c.sendChan:
			if !ok {
				return
			}
			// Broadcast to all subscribers
			c.mu.RLock()
			for _, subChan := range c.subscribers {
				select {
				case subChan <- msg:
				default:
					// Subscriber channel full, skip
				}
			}
			c.mu.RUnlock()
			
			// Also send to receive channel
			select {
			case c.recvChan <- msg:
			default:
				// Receive channel full, skip
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// ChannelManager manages multiple channels
type ChannelManager struct {
	channels map[string]*Channel
	mu       sync.RWMutex
	ctx      context.Context
}

// NewChannelManager creates a new channel manager
func NewChannelManager(ctx context.Context) *ChannelManager {
	return &ChannelManager{
		channels: make(map[string]*Channel),
		ctx:      ctx,
	}
}

// GetOrCreateChannel gets or creates a channel
func (cm *ChannelManager) GetOrCreateChannel(name string) *Channel {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if channel, ok := cm.channels[name]; ok {
		return channel
	}

	channel := NewChannel(name, cm.ctx)
	channel.Start()
	cm.channels[name] = channel
	return channel
}

// GetChannel gets a channel by name
func (cm *ChannelManager) GetChannel(name string) (*Channel, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	channel, ok := cm.channels[name]
	return channel, ok
}

// RemoveChannel removes a channel
func (cm *ChannelManager) RemoveChannel(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if channel, ok := cm.channels[name]; ok {
		channel.Stop()
		delete(cm.channels, name)
	}
}

// Close closes all channels
func (cm *ChannelManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, channel := range cm.channels {
		channel.Stop()
	}
	cm.channels = make(map[string]*Channel)
}

