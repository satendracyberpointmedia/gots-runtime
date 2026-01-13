package tsengine

// TypeInfo represents TypeScript type information
type TypeInfo struct {
	Name      string
	Kind      TypeKind
	Properties map[string]*TypeInfo
	IsOptional bool
}

// TypeKind represents the kind of TypeScript type
type TypeKind int

const (
	TypeUnknown TypeKind = iota
	TypeString
	TypeNumber
	TypeBoolean
	TypeObject
	TypeArray
	TypeFunction
	TypeAny
	TypeVoid
	TypeNull
	TypeUndefined
)

// String returns the string representation of TypeKind
func (k TypeKind) String() string {
	switch k {
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeBoolean:
		return "boolean"
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	case TypeFunction:
		return "function"
	case TypeAny:
		return "any"
	case TypeVoid:
		return "void"
	case TypeNull:
		return "null"
	case TypeUndefined:
		return "undefined"
	default:
		return "unknown"
	}
}

