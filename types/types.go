package types

import "strings"

// Type represents a type in the F1-Lang type system.
type Type interface {
	String() string
	typeNode() // marker method to ensure only types satisfy this interface
}

// Primitive types
type PrimitiveType int

const (
	Int PrimitiveType = iota
	Bool
	String
	Void
)

func (p PrimitiveType) String() string {
	switch p {
	case Int:
		return "int"
	case Bool:
		return "bool"
	case String:
		return "string"
	case Void:
		return "void"
	default:
		return "unknown"
	}
}

func (p PrimitiveType) typeNode() {}

// FunctionType represents a function type with parameters and return type.
type FunctionType struct {
	Params []Type
	Return Type
}

func (f *FunctionType) String() string {
	var params []string
	for _, p := range f.Params {
		params = append(params, p.String())
	}
	return "(" + strings.Join(params, ", ") + ") -> " + f.Return.String()
}

func (f *FunctionType) typeNode() {}

// Equal checks if two types are structurally equal.
func Equal(a, b Type) bool {
	switch at := a.(type) {
	case PrimitiveType:
		bt, ok := b.(PrimitiveType)
		return ok && at == bt
	case *FunctionType:
		bt, ok := b.(*FunctionType)
		if !ok {
			return false
		}
		if len(at.Params) != len(bt.Params) {
			return false
		}
		for i := range at.Params {
			if !Equal(at.Params[i], bt.Params[i]) {
				return false
			}
		}
		return Equal(at.Return, bt.Return)
	default:
		return false
	}
}
