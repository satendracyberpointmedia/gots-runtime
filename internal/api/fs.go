package api

import (
	"io/fs"
	"os"
	"path/filepath"

	"gots-runtime/internal/eventloop"
)

// FS provides file system operations
type FS struct {
	eventLoop *eventloop.Loop
}

// NewFS creates a new file system API
func NewFS(eventLoop *eventloop.Loop) *FS {
	return &FS{
		eventLoop: eventLoop,
	}
}

// ReadFile reads a file asynchronously
func (fs *FS) ReadFile(path string, callback func([]byte, error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		data, err := os.ReadFile(path)
		callback(data, err)
		return nil
	}, 0))
}

// WriteFile writes data to a file asynchronously
func (fs *FS) WriteFile(path string, data []byte, perm os.FileMode, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.WriteFile(path, data, perm)
		callback(err)
		return nil
	}, 0))
}

// ReadDir reads a directory asynchronously
func (fs *FS) ReadDir(path string, callback func([]fs.DirEntry, error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		entries, err := os.ReadDir(path)
		callback(entries, err)
		return nil
	}, 0))
}

// Stat gets file information asynchronously
func (fs *FS) Stat(path string, callback func(os.FileInfo, error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		info, err := os.Stat(path)
		callback(info, err)
		return nil
	}, 0))
}

// Mkdir creates a directory asynchronously
func (fs *FS) Mkdir(path string, perm os.FileMode, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.Mkdir(path, perm)
		callback(err)
		return nil
	}, 0))
}

// MkdirAll creates a directory and all parent directories asynchronously
func (fs *FS) MkdirAll(path string, perm os.FileMode, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.MkdirAll(path, perm)
		callback(err)
		return nil
	}, 0))
}

// Remove removes a file or directory asynchronously
func (fs *FS) Remove(path string, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.Remove(path)
		callback(err)
		return nil
	}, 0))
}

// RemoveAll removes a path and all children asynchronously
func (fs *FS) RemoveAll(path string, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.RemoveAll(path)
		callback(err)
		return nil
	}, 0))
}

// Rename renames a file or directory asynchronously
func (fs *FS) Rename(oldpath, newpath string, callback func(error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := os.Rename(oldpath, newpath)
		callback(err)
		return nil
	}, 0))
}

// Abs returns an absolute representation of a path
func (fs *FS) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// Join joins path elements
func (fs *FS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Base returns the last element of a path
func (fs *FS) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of a path
func (fs *FS) Dir(path string) string {
	return filepath.Dir(path)
}

// Ext returns the file name extension
func (fs *FS) Ext(path string) string {
	return filepath.Ext(path)
}

// Exists checks if a path exists
func (fs *FS) Exists(path string, callback func(bool, error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		_, err := os.Stat(path)
		if err == nil {
			callback(true, nil)
		} else if os.IsNotExist(err) {
			callback(false, nil)
		} else {
			callback(false, err)
		}
		return nil
	}, 0))
}

// ReadFileSync reads a file synchronously (for compatibility)
func (fs *FS) ReadFileSync(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFileSync writes a file synchronously (for compatibility)
func (fs *FS) WriteFileSync(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// StatSync gets file information synchronously (for compatibility)
func (fs *FS) StatSync(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// FileHandle represents an open file
type FileHandle struct {
	file *os.File
	fs   *FS
}

// Open opens a file for reading or writing
func (fs *FS) Open(path string, flag int, perm os.FileMode, callback func(*FileHandle, error)) {
	fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		file, err := os.OpenFile(path, flag, perm)
		if err != nil {
			callback(nil, err)
			return nil
		}
		callback(&FileHandle{file: file, fs: fs}, nil)
		return nil
	}, 0))
}

// Read reads from the file
func (fh *FileHandle) Read(b []byte, callback func(int, error)) {
	fh.fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		n, err := fh.file.Read(b)
		callback(n, err)
		return nil
	}, 0))
}

// Write writes to the file
func (fh *FileHandle) Write(b []byte, callback func(int, error)) {
	fh.fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		n, err := fh.file.Write(b)
		callback(n, err)
		return nil
	}, 0))
}

// Close closes the file
func (fh *FileHandle) Close(callback func(error)) {
	fh.fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		err := fh.file.Close()
		callback(err)
		return nil
	}, 0))
}

// Seek sets the offset for the next read or write
func (fh *FileHandle) Seek(offset int64, whence int, callback func(int64, error)) {
	fh.fs.eventLoop.Enqueue(eventloop.NewEvent(eventloop.EventIO, func() error {
		pos, err := fh.file.Seek(offset, whence)
		callback(pos, err)
		return nil
	}, 0))
}

