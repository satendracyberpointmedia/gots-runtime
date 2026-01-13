package api

import (
	"net"
	"time"

	"gots-runtime/internal/eventloop"
)

// Net provides network operations
type Net struct {
	eventLoop *eventloop.Loop
}

// NewNet creates a new network API
func NewNet(eventLoop *eventloop.Loop) *Net {
	return &Net{
		eventLoop: eventLoop,
	}
}

// Dial connects to a network address
func (n *Net) Dial(network, address string, callback func(net.Conn, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		conn, err := net.Dial(network, address)
		callback(conn, err)
		return nil
	}, 0))
}

// DialTimeout connects to a network address with a timeout
func (n *Net) DialTimeout(network, address string, timeout time.Duration, callback func(net.Conn, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		conn, err := net.DialTimeout(network, address, timeout)
		callback(conn, err)
		return nil
	}, 0))
}

// Listen creates a listener on a network address
func (n *Net) Listen(network, address string, callback func(net.Listener, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		listener, err := net.Listen(network, address)
		callback(listener, err)
		return nil
	}, 0))
}

// Conn represents a network connection
type Conn struct {
	conn      net.Conn
	eventLoop *eventloop.Loop
}

// NewConn wraps a net.Conn with async operations
func NewConn(conn net.Conn, eventLoop *eventloop.Loop) *Conn {
	return &Conn{
		conn:      conn,
		eventLoop: eventLoop,
	}
}

// Read reads data from the connection
func (c *Conn) Read(b []byte, callback func(int, error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		n, err := c.conn.Read(b)
		callback(n, err)
		return nil
	}, 0))
}

// Write writes data to the connection
func (c *Conn) Write(b []byte, callback func(int, error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		n, err := c.conn.Write(b)
		callback(n, err)
		return nil
	}, 0))
}

// Close closes the connection
func (c *Conn) Close(callback func(error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := c.conn.Close()
		callback(err)
		return nil
	}, 0))
}

// LocalAddr returns the local network address
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines
func (c *Conn) SetDeadline(t time.Time, callback func(error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := c.conn.SetDeadline(t)
		callback(err)
		return nil
	}, 0))
}

// SetReadDeadline sets the read deadline
func (c *Conn) SetReadDeadline(t time.Time, callback func(error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := c.conn.SetReadDeadline(t)
		callback(err)
		return nil
	}, 0))
}

// SetWriteDeadline sets the write deadline
func (c *Conn) SetWriteDeadline(t time.Time, callback func(error)) {
	c.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := c.conn.SetWriteDeadline(t)
		callback(err)
		return nil
	}, 0))
}

// Listener represents a network listener
type Listener struct {
	listener  net.Listener
	eventLoop *eventloop.Loop
}

// NewListener wraps a net.Listener with async operations
func NewListener(listener net.Listener, eventLoop *eventloop.Loop) *Listener {
	return &Listener{
		listener:  listener,
		eventLoop: eventLoop,
	}
}

// Accept accepts a connection
func (l *Listener) Accept(callback func(*Conn, error)) {
	l.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		conn, err := l.listener.Accept()
		if err != nil {
			callback(nil, err)
			return nil
		}
		callback(NewConn(conn, l.eventLoop), nil)
		return nil
	}, 0))
}

// Close closes the listener
func (l *Listener) Close(callback func(error)) {
	l.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := l.listener.Close()
		callback(err)
		return nil
	}, 0))
}

// Addr returns the listener's network address
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// ResolveTCPAddr resolves a TCP address
func (n *Net) ResolveTCPAddr(network, address string, callback func(*net.TCPAddr, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		addr, err := net.ResolveTCPAddr(network, address)
		callback(addr, err)
		return nil
	}, 0))
}

// ResolveUDPAddr resolves a UDP address
func (n *Net) ResolveUDPAddr(network, address string, callback func(*net.UDPAddr, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		addr, err := net.ResolveUDPAddr(network, address)
		callback(addr, err)
		return nil
	}, 0))
}

// LookupIP looks up IP addresses for a hostname
func (n *Net) LookupIP(host string, callback func([]net.IP, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		ips, err := net.LookupIP(host)
		callback(ips, err)
		return nil
	}, 0))
}

// LookupHost looks up host addresses for a hostname
func (n *Net) LookupHost(host string, callback func([]string, error)) {
	n.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		addrs, err := net.LookupHost(host)
		callback(addrs, err)
		return nil
	}, 0))
}

