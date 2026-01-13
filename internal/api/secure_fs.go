package api

import (
	"io/fs"
	"os"

	"gots-runtime/internal/eventloop"
	"gots-runtime/internal/security"
)

// SecureFS provides file system operations with security
type SecureFS struct {
	fs          *FS
	permManager *security.PermissionManager
	moduleID    string
}

// NewSecureFS creates a new secure file system API
func NewSecureFS(eventLoop *eventloop.Loop, permManager *security.PermissionManager, moduleID string) *SecureFS {
	return &SecureFS{
		fs:          NewFS(eventLoop),
		permManager: permManager,
		moduleID:    moduleID,
	}
}

// ReadFile reads a file asynchronously with permission check
func (sfs *SecureFS) ReadFile(path string, callback func([]byte, error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSRead); err != nil {
		callback(nil, err)
		return
	}
	
	sfs.fs.ReadFile(path, callback)
}

// WriteFile writes data to a file asynchronously with permission check
func (sfs *SecureFS) WriteFile(path string, data []byte, perm os.FileMode, callback func(error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSWrite); err != nil {
		callback(err)
		return
	}
	
	sfs.fs.WriteFile(path, data, perm, callback)
}

// ReadDir reads a directory asynchronously with permission check
func (sfs *SecureFS) ReadDir(path string, callback func([]fs.DirEntry, error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSRead); err != nil {
		callback(nil, err)
		return
	}
	
	sfs.fs.ReadDir(path, callback)
}

// Stat gets file information asynchronously with permission check
func (sfs *SecureFS) Stat(path string, callback func(os.FileInfo, error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSRead); err != nil {
		callback(nil, err)
		return
	}
	
	sfs.fs.Stat(path, callback)
}

// Mkdir creates a directory asynchronously with permission check
func (sfs *SecureFS) Mkdir(path string, perm os.FileMode, callback func(error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSWrite); err != nil {
		callback(err)
		return
	}
	
	sfs.fs.Mkdir(path, perm, callback)
}

// Remove removes a file or directory asynchronously with permission check
func (sfs *SecureFS) Remove(path string, callback func(error)) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSWrite); err != nil {
		callback(err)
		return
	}
	
	sfs.fs.Remove(path, callback)
}

// Open opens a file for reading or writing with permission check
func (sfs *SecureFS) Open(path string, flag int, perm os.FileMode, callback func(*FileHandle, error)) {
	// Determine permission based on flag
	var permType security.Permission
	if flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0 {
		permType = security.PermissionFSWrite
	} else {
		permType = security.PermissionFSRead
	}
	
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, permType); err != nil {
		callback(nil, err)
		return
	}
	
	sfs.fs.Open(path, flag, perm, callback)
}

// ReadFileSync reads a file synchronously with permission check
func (sfs *SecureFS) ReadFileSync(path string) ([]byte, error) {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSRead); err != nil {
		return nil, err
	}
	
	return sfs.fs.ReadFileSync(path)
}

// WriteFileSync writes a file synchronously with permission check
func (sfs *SecureFS) WriteFileSync(path string, data []byte, perm os.FileMode) error {
	// Check permission
	if err := sfs.permManager.CheckPermission(sfs.moduleID, security.PermissionFSWrite); err != nil {
		return err
	}
	
	return sfs.fs.WriteFileSync(path, data, perm)
}

