package typechecker

import "f1c/types"

// Environment stores type bindings for variables and functions.
type Environment struct {
	store map[string]types.Type
	outer *Environment
}

// NewEnvironment creates a new top-level environment.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]types.Type),
		outer: nil,
	}
}

// NewEnclosedEnvironment creates a new environment enclosed by an outer scope.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]types.Type),
		outer: outer,
	}
}

// Get retrieves a type binding, searching outer scopes if necessary.
func (e *Environment) Get(name string) (types.Type, bool) {
	typ, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return typ, ok
}

// Set creates or updates a type binding in the current scope.
func (e *Environment) Set(name string, typ types.Type) {
	e.store[name] = typ
}

// ExistsInCurrentScope checks if a name exists in the current scope only.
func (e *Environment) ExistsInCurrentScope(name string) bool {
	_, ok := e.store[name]
	return ok
}
