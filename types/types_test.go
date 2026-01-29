package types

import "testing"

func TestPrimitiveTypes_String(t *testing.T) {
	tests := []struct {
		typ      Type
		expected string
	}{
		{Int, "int"},
		{Bool, "bool"},
		{String, "string"},
		{Void, "void"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFunctionType_String(t *testing.T) {
	tests := []struct {
		name     string
		fn       *FunctionType
		expected string
	}{
		{
			name:     "no params void return",
			fn:       &FunctionType{Params: []Type{}, Return: Void},
			expected: "() -> void",
		},
		{
			name:     "one param int return",
			fn:       &FunctionType{Params: []Type{Int}, Return: Int},
			expected: "(int) -> int",
		},
		{
			name:     "two params bool return",
			fn:       &FunctionType{Params: []Type{Int, Int}, Return: Bool},
			expected: "(int, int) -> bool",
		},
		{
			name:     "mixed params",
			fn:       &FunctionType{Params: []Type{Int, String, Bool}, Return: String},
			expected: "(int, string, bool) -> string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFunctionType_IsType(t *testing.T) {
	fn := &FunctionType{Params: []Type{Int}, Return: Int}
	// FunctionType should satisfy Type interface
	var _ Type = fn
}

func TestTypeEquality(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Type
		expected bool
	}{
		{"int == int", Int, Int, true},
		{"bool == bool", Bool, Bool, true},
		{"string == string", String, String, true},
		{"void == void", Void, Void, true},
		{"int != bool", Int, Bool, false},
		{"int != string", Int, String, false},
		{
			"func(int)->int == func(int)->int",
			&FunctionType{Params: []Type{Int}, Return: Int},
			&FunctionType{Params: []Type{Int}, Return: Int},
			true,
		},
		{
			"func(int)->int != func(int)->bool",
			&FunctionType{Params: []Type{Int}, Return: Int},
			&FunctionType{Params: []Type{Int}, Return: Bool},
			false,
		},
		{
			"func(int)->int != func(bool)->int",
			&FunctionType{Params: []Type{Int}, Return: Int},
			&FunctionType{Params: []Type{Bool}, Return: Int},
			false,
		},
		{
			"func(int,int)->int != func(int)->int",
			&FunctionType{Params: []Type{Int, Int}, Return: Int},
			&FunctionType{Params: []Type{Int}, Return: Int},
			false,
		},
		{
			"int != func()->int",
			Int,
			&FunctionType{Params: []Type{}, Return: Int},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.a, tt.b); got != tt.expected {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}
