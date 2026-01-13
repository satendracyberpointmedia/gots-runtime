package tsengine

import (
	"fmt"
	"reflect"
	"sync"
)

// TypeValidator validates types at runtime
type TypeValidator struct {
	hooks map[string]TypeValidationHook
	mu    sync.RWMutex
}

// TypeValidationHook is a function that validates a value against a type
type TypeValidationHook func(value interface{}, expectedType *TypeInfo) error

// NewTypeValidator creates a new type validator
func NewTypeValidator() *TypeValidator {
	return &TypeValidator{
		hooks: make(map[string]TypeValidationHook),
	}
}

// RegisterHook registers a type validation hook
func (tv *TypeValidator) RegisterHook(typeName string, hook TypeValidationHook) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	tv.hooks[typeName] = hook
}

// Validate validates a value against a type
func (tv *TypeValidator) Validate(value interface{}, expectedType *TypeInfo) error {
	if expectedType == nil {
		return nil // No type to validate against
	}
	
	// Check for custom hook
	tv.mu.RLock()
	hook, hasHook := tv.hooks[expectedType.Name]
	tv.mu.RUnlock()
	
	if hasHook {
		return hook(value, expectedType)
	}
	
	// Default validation based on type kind
	return tv.validateByKind(value, expectedType)
}

// validateByKind validates based on TypeKind
func (tv *TypeValidator) validateByKind(value interface{}, typeInfo *TypeInfo) error {
	switch typeInfo.Kind {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case TypeNumber:
		if !isNumber(value) {
			return fmt.Errorf("expected number, got %T", value)
		}
	case TypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case TypeObject:
		if !isObject(value) {
			return fmt.Errorf("expected object, got %T", value)
		}
		// Validate properties if type info has them
		if typeInfo.Properties != nil {
			return tv.validateObjectProperties(value, typeInfo)
		}
	case TypeArray:
		if !isArray(value) {
			return fmt.Errorf("expected array, got %T", value)
		}
	case TypeFunction:
		if !isFunction(value) {
			return fmt.Errorf("expected function, got %T", value)
		}
	case TypeAny:
		return nil // Any type is valid
	case TypeVoid:
		if value != nil {
			return fmt.Errorf("expected void, got %T", value)
		}
	case TypeNull:
		if value != nil {
			return fmt.Errorf("expected null, got %T", value)
		}
	case TypeUndefined:
		// Undefined is typically nil in Go
		return nil
	default:
		return fmt.Errorf("unknown type kind: %s", typeInfo.Kind)
	}
	
	return nil
}

// validateObjectProperties validates object properties
func (tv *TypeValidator) validateObjectProperties(value interface{}, typeInfo *TypeInfo) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map && val.Kind() != reflect.Struct {
		return fmt.Errorf("cannot validate properties of %T", value)
	}
	
	// For maps, check keys
	if val.Kind() == reflect.Map {
		for key, propType := range typeInfo.Properties {
			mapVal := val.MapIndex(reflect.ValueOf(key))
			if !mapVal.IsValid() {
				if !propType.IsOptional {
					return fmt.Errorf("missing required property: %s", key)
				}
				continue
			}
			
			if err := tv.Validate(mapVal.Interface(), propType); err != nil {
				return fmt.Errorf("property %s: %w", key, err)
			}
		}
	}
	
	return nil
}

// Helper functions
func isNumber(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	}
	return false
}

func isObject(v interface{}) bool {
	if v == nil {
		return false
	}
	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Map || kind == reflect.Struct
}

func isArray(v interface{}) bool {
	if v == nil {
		return false
	}
	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func isFunction(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Func
}

// TypeEnforcer enforces types at runtime
type TypeEnforcer struct {
	validator *TypeValidator
	enabled   bool
	mu        sync.RWMutex
}

// NewTypeEnforcer creates a new type enforcer
func NewTypeEnforcer() *TypeEnforcer {
	return &TypeEnforcer{
		validator: NewTypeValidator(),
		enabled:   true,
	}
}

// Enable enables type enforcement
func (te *TypeEnforcer) Enable() {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.enabled = true
}

// Disable disables type enforcement
func (te *TypeEnforcer) Disable() {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.enabled = false
}

// Enforce enforces a type on a value
func (te *TypeEnforcer) Enforce(value interface{}, expectedType *TypeInfo) error {
	te.mu.RLock()
	enabled := te.enabled
	te.mu.RUnlock()
	
	if !enabled {
		return nil
	}
	
	return te.validator.Validate(value, expectedType)
}

// RegisterHook registers a validation hook
func (te *TypeEnforcer) RegisterHook(typeName string, hook TypeValidationHook) {
	te.validator.RegisterHook(typeName, hook)
}

