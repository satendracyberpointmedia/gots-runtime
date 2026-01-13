package api

import (
	"net"
	"time"

	"gots-runtime/internal/eventloop"
	"gots-runtime/internal/security"
)

// SecureNet provides network operations with security
type SecureNet struct {
	net         *Net
	permManager *security.PermissionManager
	moduleID    string
}

// NewSecureNet creates a new secure network API
func NewSecureNet(eventLoop *eventloop.Loop, permManager *security.PermissionManager, moduleID string) *SecureNet {
	return &SecureNet{
		net:         NewNet(eventLoop),
		permManager: permManager,
		moduleID:    moduleID,
	}
}

// Dial connects to a network address with permission check
func (sn *SecureNet) Dial(network, address string, callback func(net.Conn, error)) {
	// Check permission
	if err := sn.permManager.CheckPermission(sn.moduleID, security.PermissionNetDial); err != nil {
		callback(nil, err)
		return
	}
	
	sn.net.Dial(network, address, callback)
}

// DialTimeout connects to a network address with timeout and permission check
func (sn *SecureNet) DialTimeout(network, address string, timeout time.Duration, callback func(net.Conn, error)) {
	// Check permission
	if err := sn.permManager.CheckPermission(sn.moduleID, security.PermissionNetDial); err != nil {
		callback(nil, err)
		return
	}
	
	sn.net.DialTimeout(network, address, timeout, callback)
}

// Listen creates a listener on a network address with permission check
func (sn *SecureNet) Listen(network, address string, callback func(net.Listener, error)) {
	// Check permission
	if err := sn.permManager.CheckPermission(sn.moduleID, security.PermissionNetListen); err != nil {
		callback(nil, err)
		return
	}
	
	sn.net.Listen(network, address, callback)
}

// LookupIP looks up IP addresses for a hostname with permission check
func (sn *SecureNet) LookupIP(host string, callback func([]net.IP, error)) {
	// Check permission (DNS lookup requires net permission)
	if err := sn.permManager.CheckPermission(sn.moduleID, security.PermissionNetDial); err != nil {
		callback(nil, err)
		return
	}
	
	sn.net.LookupIP(host, callback)
}

// LookupHost looks up host addresses for a hostname with permission check
func (sn *SecureNet) LookupHost(host string, callback func([]string, error)) {
	// Check permission
	if err := sn.permManager.CheckPermission(sn.moduleID, security.PermissionNetDial); err != nil {
		callback(nil, err)
		return
	}
	
	sn.net.LookupHost(host, callback)
}

