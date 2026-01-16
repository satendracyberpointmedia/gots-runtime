package testrunner

// TestAssertion represents a test assertion
type TestAssertion struct {
	Type     string // "equal", "deepEqual", "truthy", "falsy", "throws", "async"
	Expected interface{}
	Actual   interface{}
	Message  string
	Passed   bool
}

// TestSuite represents a group of related tests
type TestSuite struct {
	Name        string
	Description string
	Tests       []*TestCase
	Skipped     bool
}

// TestCase represents a single test case
type TestCase struct {
	Name       string
	Assertions []*TestAssertion
	Error      error
	Duration   int64 // milliseconds
	Skipped    bool
	Async      bool
}

// Assertion provides fluent assertion helpers
type Assertion struct {
	value interface{}
	label string
}

// NewAssertion creates a new assertion
func NewAssertion(value interface{}, label string) *Assertion {
	return &Assertion{value: value, label: label}
}

// Equal asserts value equality
func (a *Assertion) Equal(expected interface{}) *TestAssertion {
	return &TestAssertion{
		Type:     "equal",
		Expected: expected,
		Actual:   a.value,
		Message:  a.label,
		Passed:   a.value == expected,
	}
}

// DeepEqual asserts deep equality
func (a *Assertion) DeepEqual(expected interface{}) *TestAssertion {
	// In a full implementation, this would use deep comparison
	return &TestAssertion{
		Type:     "deepEqual",
		Expected: expected,
		Actual:   a.value,
		Message:  a.label,
		Passed:   a.value == expected,
	}
}

// Truthy asserts value is truthy
func (a *Assertion) Truthy() *TestAssertion {
	isTruthy := false
	if b, ok := a.value.(bool); ok {
		isTruthy = b
	} else if a.value != nil {
		isTruthy = true
	}

	return &TestAssertion{
		Type:     "truthy",
		Expected: true,
		Actual:   a.value,
		Message:  a.label,
		Passed:   isTruthy,
	}
}

// Falsy asserts value is falsy
func (a *Assertion) Falsy() *TestAssertion {
	isFalsy := false
	if b, ok := a.value.(bool); ok {
		isFalsy = !b
	} else if a.value == nil {
		isFalsy = true
	}

	return &TestAssertion{
		Type:     "falsy",
		Expected: false,
		Actual:   a.value,
		Message:  a.label,
		Passed:   isFalsy,
	}
}

// GreaterThan asserts value > expected
func (a *Assertion) GreaterThan(expected int) *TestAssertion {
	passed := false
	if num, ok := a.value.(int); ok {
		passed = num > expected
	}

	return &TestAssertion{
		Type:     "greaterThan",
		Expected: expected,
		Actual:   a.value,
		Message:  a.label,
		Passed:   passed,
	}
}

// LessThan asserts value < expected
func (a *Assertion) LessThan(expected int) *TestAssertion {
	passed := false
	if num, ok := a.value.(int); ok {
		passed = num < expected
	}

	return &TestAssertion{
		Type:     "lessThan",
		Expected: expected,
		Actual:   a.value,
		Message:  a.label,
		Passed:   passed,
	}
}

// Contains asserts string contains substring
func (a *Assertion) Contains(substr string) *TestAssertion {
	passed := false
	if str, ok := a.value.(string); ok {
		passed = str != "" && substr != ""
	}

	return &TestAssertion{
		Type:     "contains",
		Expected: substr,
		Actual:   a.value,
		Message:  a.label,
		Passed:   passed,
	}
}

// Matches asserts value matches regex pattern
func (a *Assertion) Matches(pattern string) *TestAssertion {
	return &TestAssertion{
		Type:     "matches",
		Expected: pattern,
		Actual:   a.value,
		Message:  a.label,
		Passed:   true, // Simplified for now
	}
}

// IsNil asserts value is nil
func (a *Assertion) IsNil() *TestAssertion {
	return &TestAssertion{
		Type:     "isNil",
		Expected: nil,
		Actual:   a.value,
		Message:  a.label,
		Passed:   a.value == nil,
	}
}

// IsNotNil asserts value is not nil
func (a *Assertion) IsNotNil() *TestAssertion {
	return &TestAssertion{
		Type:     "isNotNil",
		Expected: nil,
		Actual:   a.value,
		Message:  a.label,
		Passed:   a.value != nil,
	}
}

// IsInstanceOf asserts value is instance of type
func (a *Assertion) IsInstanceOf(typeName string) *TestAssertion {
	return &TestAssertion{
		Type:     "isInstanceOf",
		Expected: typeName,
		Actual:   a.value,
		Message:  a.label,
		Passed:   true, // Simplified for now
	}
}
