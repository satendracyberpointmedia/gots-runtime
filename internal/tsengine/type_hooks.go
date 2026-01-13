package tsengine

import (
	"fmt"
	"reflect"
)

// RuntimeTypeHooks provides runtime-level type validation hooks
type RuntimeTypeHooks struct {
	enforcer *TypeEnforcer
}

// NewRuntimeTypeHooks creates new runtime type hooks
func NewRuntimeTypeHooks(enforcer *TypeEnforcer) *RuntimeTypeHooks {
	return &RuntimeTypeHooks{
		enforcer: enforcer,
	}
}

// RegisterDefaultHooks registers default type validation hooks
func (rth *RuntimeTypeHooks) RegisterDefaultHooks() {
	// Register hooks for common types
	rth.enforcer.RegisterHook("string", rth.validateString)
	rth.enforcer.RegisterHook("number", rth.validateNumber)
	rth.enforcer.RegisterHook("boolean", rth.validateBoolean)
	rth.enforcer.RegisterHook("object", rth.validateObject)
	rth.enforcer.RegisterHook("array", rth.validateArray)
}

// validateString validates a string type
func (rth *RuntimeTypeHooks) validateString(value interface{}, expectedType *TypeInfo) error {
	if str, ok := value.(string); ok {
		// Additional validation could go here
		_ = str
		return nil
	}
	return fmt.Errorf("type mismatch: expected string, got %T", value)
}

// validateNumber validates a number type
func (rth *RuntimeTypeHooks) validateNumber(value interface{}, expectedType *TypeInfo) error {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil
	}
	return fmt.Errorf("type mismatch: expected number, got %T", value)
}

// validateBoolean validates a boolean type
func (rth *RuntimeTypeHooks) validateBoolean(value interface{}, expectedType *TypeInfo) error {
	if _, ok := value.(bool); ok {
		return nil
	}
	return fmt.Errorf("type mismatch: expected boolean, got %T", value)
}

// validateObject validates an object type
func (rth *RuntimeTypeHooks) validateObject(value interface{}, expectedType *TypeInfo) error {
	if value == nil {
		return fmt.Errorf("expected object, got nil")
	}
	
	val := reflect.ValueOf(value)
	kind := val.Kind()
	
	if kind != reflect.Map && kind != reflect.Struct {
		return fmt.Errorf("type mismatch: expected object, got %T", value)
	}
	
	// Validate properties if type info has them
	if expectedType.Properties != nil {
		return rth.validateObjectProperties(value, expectedType)
	}
	
	return nil
}

// validateArray validates an array type
func (rth *RuntimeTypeHooks) validateArray(value interface{}, expectedType *TypeInfo) error {
	if value == nil {
		return fmt.Errorf("expected array, got nil")
	}
	
	val := reflect.ValueOf(value)
	kind := val.Kind()
	
	if kind != reflect.Slice && kind != reflect.Array {
		return fmt.Errorf("type mismatch: expected array, got %T", value)
	}
	
	return nil
}

// validateObjectProperties validates object properties
func (rth *RuntimeTypeHooks) validateObjectProperties(value interface{}, typeInfo *TypeInfo) error {
	val := reflect.ValueOf(value)
	
	if val.Kind() == reflect.Map {
		for key, propType := range typeInfo.Properties {
			keyVal := reflect.ValueOf(key)
			mapVal := val.MapIndex(keyVal)
			
			if !mapVal.IsValid() {
				if !propType.IsOptional {
					return fmt.Errorf("missing required property: %s", key)
				}
				continue
			}
			
			if err := rth.enforcer.Enforce(mapVal.Interface(), propType); err != nil {
				return fmt.Errorf("property %s: %w", key, err)
			}
		}
	}
	
	return nil
}

// FunctionCallValidator validates function calls
type FunctionCallValidator struct {
	enforcer *TypeEnforcer
}

// NewFunctionCallValidator creates a new function call validator
func NewFunctionCallValidator(enforcer *TypeEnforcer) *FunctionCallValidator {
	return &FunctionCallValidator{
		enforcer: enforcer,
	}
}

// ValidateCall validates function call arguments and return value
func (fcv *FunctionCallValidator) ValidateCall(args []interface{}, argTypes []*TypeInfo, returnValue interface{}, returnType *TypeInfo) error {
	// Validate arguments
	if len(args) != len(argTypes) {
		return fmt.Errorf("argument count mismatch: expected %d, got %d", len(argTypes), len(args))
	}
	
	for i, arg := range args {
		if i < len(argTypes) {
			if err := fcv.enforcer.Enforce(arg, argTypes[i]); err != nil {
				return fmt.Errorf("argument %d: %w", i, err)
			}
		}
	}
	
	// Validate return value
	if returnType != nil {
		if err := fcv.enforcer.Enforce(returnValue, returnType); err != nil {
			return fmt.Errorf("return value: %w", err)
		}
	}
	
	return nil
}

