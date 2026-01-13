package data

import (
	"fmt"
	"sync"
)

// ImmutableMap is an immutable map
type ImmutableMap struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewImmutableMap creates a new immutable map
func NewImmutableMap() *ImmutableMap {
	return &ImmutableMap{
		data: make(map[string]interface{}),
	}
}

// Get gets a value
func (im *ImmutableMap) Get(key string) (interface{}, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	value, ok := im.data[key]
	return value, ok
}

// Set sets a value (returns a new map)
func (im *ImmutableMap) Set(key string, value interface{}) *ImmutableMap {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	// Create new map with all existing data
	newData := make(map[string]interface{})
	for k, v := range im.data {
		newData[k] = v
	}
	
	// Add new value
	newData[key] = value
	
	return &ImmutableMap{
		data: newData,
	}
}

// Delete deletes a key (returns a new map)
func (im *ImmutableMap) Delete(key string) *ImmutableMap {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	// Create new map without the key
	newData := make(map[string]interface{})
	for k, v := range im.data {
		if k != key {
			newData[k] = v
		}
	}
	
	return &ImmutableMap{
		data: newData,
	}
}

// Size returns the size of the map
func (im *ImmutableMap) Size() int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return len(im.data)
}

// Keys returns all keys
func (im *ImmutableMap) Keys() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	keys := make([]string, 0, len(im.data))
	for k := range im.data {
		keys = append(keys, k)
	}
	return keys
}

// ImmutableList is an immutable list
type ImmutableList struct {
	data []interface{}
	mu   sync.RWMutex
}

// NewImmutableList creates a new immutable list
func NewImmutableList() *ImmutableList {
	return &ImmutableList{
		data: make([]interface{}, 0),
	}
}

// Get gets a value at index
func (il *ImmutableList) Get(index int) (interface{}, error) {
	il.mu.RLock()
	defer il.mu.RUnlock()
	
	if index < 0 || index >= len(il.data) {
		return nil, fmt.Errorf("index out of bounds: %d", index)
	}
	
	return il.data[index], nil
}

// Append appends a value (returns a new list)
func (il *ImmutableList) Append(value interface{}) *ImmutableList {
	il.mu.RLock()
	defer il.mu.RUnlock()
	
	// Create new list with all existing data
	newData := make([]interface{}, len(il.data), len(il.data)+1)
	copy(newData, il.data)
	newData = append(newData, value)
	
	return &ImmutableList{
		data: newData,
	}
}

// Prepend prepends a value (returns a new list)
func (il *ImmutableList) Prepend(value interface{}) *ImmutableList {
	il.mu.RLock()
	defer il.mu.RUnlock()
	
	// Create new list with new value at front
	newData := make([]interface{}, 0, len(il.data)+1)
	newData = append(newData, value)
	newData = append(newData, il.data...)
	
	return &ImmutableList{
		data: newData,
	}
}

// Size returns the size of the list
func (il *ImmutableList) Size() int {
	il.mu.RLock()
	defer il.mu.RUnlock()
	return len(il.data)
}

// ImmutableSet is an immutable set
type ImmutableSet struct {
	data map[interface{}]bool
	mu   sync.RWMutex
}

// NewImmutableSet creates a new immutable set
func NewImmutableSet() *ImmutableSet {
	return &ImmutableSet{
		data: make(map[interface{}]bool),
	}
}

// Contains checks if a value is in the set
func (is *ImmutableSet) Contains(value interface{}) bool {
	is.mu.RLock()
	defer is.mu.RUnlock()
	return is.data[value]
}

// Add adds a value (returns a new set)
func (is *ImmutableSet) Add(value interface{}) *ImmutableSet {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	// Create new set with all existing data
	newData := make(map[interface{}]bool)
	for k := range is.data {
		newData[k] = true
	}
	
	// Add new value
	newData[value] = true
	
	return &ImmutableSet{
		data: newData,
	}
}

// Remove removes a value (returns a new set)
func (is *ImmutableSet) Remove(value interface{}) *ImmutableSet {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	// Create new set without the value
	newData := make(map[interface{}]bool)
	for k := range is.data {
		if k != value {
			newData[k] = true
		}
	}
	
	return &ImmutableSet{
		data: newData,
	}
}

// Size returns the size of the set
func (is *ImmutableSet) Size() int {
	is.mu.RLock()
	defer is.mu.RUnlock()
	return len(is.data)
}

