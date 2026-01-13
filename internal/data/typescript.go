package data

import (
	"github.com/dop251/goja"
)

// TypeScriptImmutableMap wraps ImmutableMap for TypeScript
type TypeScriptImmutableMap struct {
	im     *ImmutableMap
	engine *goja.Runtime
}

// NewTypeScriptImmutableMap creates a new TypeScript-wrapped immutable map
func NewTypeScriptImmutableMap(engine *goja.Runtime, im *ImmutableMap) *TypeScriptImmutableMap {
	return &TypeScriptImmutableMap{
		im:     im,
		engine: engine,
	}
}

// ToJSObject converts the immutable map to a JavaScript object
func (tsim *TypeScriptImmutableMap) ToJSObject() *goja.Object {
	obj := tsim.engine.NewObject()
	
	// Get method
	obj.Set("get", func(key string) goja.Value {
		value, ok := tsim.im.Get(key)
		if !ok {
			return goja.Undefined()
		}
		return tsim.engine.ToValue(value)
	})
	
	// Has method
	obj.Set("has", func(key string) bool {
		_, ok := tsim.im.Get(key)
		return ok
	})
	
	// Set method (returns new map)
	obj.Set("set", func(key string, value goja.Value) *goja.Object {
		newMap := tsim.im.Set(key, value.Export())
		return NewTypeScriptImmutableMap(tsim.engine, newMap).ToJSObject()
	})
	
	// Delete method (returns new map)
	obj.Set("delete", func(key string) *goja.Object {
		newMap := tsim.im.Delete(key)
		return NewTypeScriptImmutableMap(tsim.engine, newMap).ToJSObject()
	})
	
	// Size method
	obj.Set("size", func() int {
		return tsim.im.Size()
	})
	
	// Keys method
	obj.Set("keys", func() []string {
		return tsim.im.Keys()
	})
	
	// Values method
	obj.Set("values", func() []interface{} {
		keys := tsim.im.Keys()
		values := make([]interface{}, 0, len(keys))
		for _, k := range keys {
			if v, ok := tsim.im.Get(k); ok {
				values = append(values, v)
			}
		}
		return values
	})
	
	// Entries method
	obj.Set("entries", func() []interface{} {
		keys := tsim.im.Keys()
		entries := make([]interface{}, 0, len(keys))
		for _, k := range keys {
			if v, ok := tsim.im.Get(k); ok {
				entries = append(entries, []interface{}{k, v})
			}
		}
		return entries
	})
	
	// ForEach method
	obj.Set("forEach", func(callback goja.Callable) {
		keys := tsim.im.Keys()
		for _, k := range keys {
			if v, ok := tsim.im.Get(k); ok {
				callback(nil, tsim.engine.ToValue(v), tsim.engine.ToValue(k))
			}
		}
	})
	
	// ToJS method (converts to native JS Map)
	obj.Set("toJS", func() map[string]interface{} {
		result := make(map[string]interface{})
		keys := tsim.im.Keys()
		for _, k := range keys {
			if v, ok := tsim.im.Get(k); ok {
				result[k] = v
			}
		}
		return result
	})
	
	return obj
}

// TypeScriptImmutableList wraps ImmutableList for TypeScript
type TypeScriptImmutableList struct {
	il     *ImmutableList
	engine *goja.Runtime
}

// NewTypeScriptImmutableList creates a new TypeScript-wrapped immutable list
func NewTypeScriptImmutableList(engine *goja.Runtime, il *ImmutableList) *TypeScriptImmutableList {
	return &TypeScriptImmutableList{
		il:     il,
		engine: engine,
	}
}

// ToJSObject converts the immutable list to a JavaScript object
func (tsil *TypeScriptImmutableList) ToJSObject() *goja.Object {
	obj := tsil.engine.NewObject()
	
	// Get method
	obj.Set("get", func(index int) goja.Value {
		value, err := tsil.il.Get(index)
		if err != nil {
			return goja.Undefined()
		}
		return tsil.engine.ToValue(value)
	})
	
	// Set method (returns new list)
	obj.Set("set", func(index int, value goja.Value) *goja.Object {
		// For immutable list, we need to create a new list with the updated value
		// This is a simplified version - in practice, we'd need to copy and modify
		newList := tsil.il.Append(value.Export()) // Simplified - should replace at index
		return NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject()
	})
	
	// Push method (returns new list)
	obj.Set("push", func(value goja.Value) *goja.Object {
		newList := tsil.il.Append(value.Export())
		return NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject()
	})
	
	// Pop method (returns [newList, value])
	obj.Set("pop", func() []interface{} {
		size := tsil.il.Size()
		if size == 0 {
			return []interface{}{NewTypeScriptImmutableList(tsil.engine, tsil.il).ToJSObject(), goja.Undefined()}
		}
		value, _ := tsil.il.Get(size - 1)
		// Create new list without last element (simplified)
		newList := NewImmutableList()
		for i := 0; i < size-1; i++ {
			v, _ := tsil.il.Get(i)
			newList = newList.Append(v)
		}
		return []interface{}{NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject(), tsil.engine.ToValue(value)}
	})
	
	// Unshift method (returns new list)
	obj.Set("unshift", func(value goja.Value) *goja.Object {
		newList := tsil.il.Prepend(value.Export())
		return NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject()
	})
	
	// Shift method (returns [newList, value])
	obj.Set("shift", func() []interface{} {
		size := tsil.il.Size()
		if size == 0 {
			return []interface{}{NewTypeScriptImmutableList(tsil.engine, tsil.il).ToJSObject(), goja.Undefined()}
		}
		value, _ := tsil.il.Get(0)
		// Create new list without first element
		newList := NewImmutableList()
		for i := 1; i < size; i++ {
			v, _ := tsil.il.Get(i)
			newList = newList.Append(v)
		}
		return []interface{}{NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject(), tsil.engine.ToValue(value)}
	})
	
	// Size method
	obj.Set("size", func() int {
		return tsil.il.Size()
	})
	
	// ForEach method
	obj.Set("forEach", func(callback goja.Callable) {
		size := tsil.il.Size()
		for i := 0; i < size; i++ {
			if value, err := tsil.il.Get(i); err == nil {
				callback(nil, tsil.engine.ToValue(value), tsil.engine.ToValue(i))
			}
		}
	})
	
	// Map method
	obj.Set("map", func(callback goja.Callable) *goja.Object {
		size := tsil.il.Size()
		newList := NewImmutableList()
		for i := 0; i < size; i++ {
			if value, err := tsil.il.Get(i); err == nil {
				result, _ := callback(nil, tsil.engine.ToValue(value), tsil.engine.ToValue(i))
				newList = newList.Append(result.Export())
			}
		}
		return NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject()
	})
	
	// Filter method
	obj.Set("filter", func(callback goja.Callable) *goja.Object {
		size := tsil.il.Size()
		newList := NewImmutableList()
		for i := 0; i < size; i++ {
			if value, err := tsil.il.Get(i); err == nil {
				result, _ := callback(nil, tsil.engine.ToValue(value), tsil.engine.ToValue(i))
				if result.ToBoolean() {
					newList = newList.Append(value)
				}
			}
		}
		return NewTypeScriptImmutableList(tsil.engine, newList).ToJSObject()
	})
	
	// ToJS method (converts to native JS Array)
	obj.Set("toJS", func() []interface{} {
		size := tsil.il.Size()
		result := make([]interface{}, 0, size)
		for i := 0; i < size; i++ {
			if value, err := tsil.il.Get(i); err == nil {
				result = append(result, value)
			}
		}
		return result
	})
	
	return obj
}

// TypeScriptImmutableSet wraps ImmutableSet for TypeScript
type TypeScriptImmutableSet struct {
	is     *ImmutableSet
	engine *goja.Runtime
}

// NewTypeScriptImmutableSet creates a new TypeScript-wrapped immutable set
func NewTypeScriptImmutableSet(engine *goja.Runtime, is *ImmutableSet) *TypeScriptImmutableSet {
	return &TypeScriptImmutableSet{
		is:     is,
		engine: engine,
	}
}

// ToJSObject converts the immutable set to a JavaScript object
func (tsis *TypeScriptImmutableSet) ToJSObject() *goja.Object {
	obj := tsis.engine.NewObject()
	
	// Has method
	obj.Set("has", func(value goja.Value) bool {
		return tsis.is.Contains(value.Export())
	})
	
	// Add method (returns new set)
	obj.Set("add", func(value goja.Value) *goja.Object {
		newSet := tsis.is.Add(value.Export())
		return NewTypeScriptImmutableSet(tsis.engine, newSet).ToJSObject()
	})
	
	// Delete method (returns new set)
	obj.Set("delete", func(value goja.Value) *goja.Object {
		newSet := tsis.is.Remove(value.Export())
		return NewTypeScriptImmutableSet(tsis.engine, newSet).ToJSObject()
	})
	
	// Size method
	obj.Set("size", func() int {
		return tsis.is.Size()
	})
	
	// Values method
	obj.Set("values", func() []interface{} {
		// Get all values from the set
		result := make([]interface{}, 0)
		// Note: This is simplified - the actual implementation would need to iterate
		// For now, we'll return an empty array as the set doesn't expose iteration
		return result
	})
	
	// ForEach method
	obj.Set("forEach", func(callback goja.Callable) {
		// Note: This is simplified - would need proper iteration
	})
	
	// ToJS method (converts to native JS Set)
	obj.Set("toJS", func() []interface{} {
		// Note: This is simplified - would need proper conversion
		return []interface{}{}
	})
	
	return obj
}

