package runtime

import "fmt"

// Environment represents a variable scope.
type Environment struct {
	store  map[string]*Value
	parent *Environment
}

// NewEnvironment creates a new environment with an optional parent scope.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		store:  make(map[string]*Value),
		parent: parent,
	}
}

// Get looks up a variable by name, walking up the scope chain.
func (e *Environment) Get(name string) (*Value, error) {
	if val, ok := e.store[name]; ok {
		return val, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("'%s' is not defined", name)
}

// Set assigns a value to a variable in the current scope.
func (e *Environment) Set(name string, val *Value) {
	e.store[name] = val
}

// Update assigns a value to an existing variable, walking up the scope chain.
// Returns an error if the variable is not found in any scope.
func (e *Environment) Update(name string, val *Value) error {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return nil
	}
	if e.parent != nil {
		return e.parent.Update(name, val)
	}
	return fmt.Errorf("'%s' is not defined — use 'store' to declare it first", name)
}

// Has checks if a variable exists in the current scope (not parents).
func (e *Environment) Has(name string) bool {
	_, ok := e.store[name]
	return ok
}
