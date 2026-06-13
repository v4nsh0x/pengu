package runtime

import (
	"fmt"
	"strings"
	"sync"

	"github.com/v4nsh0x/pengu/ast"
)

// ValueType represents the type of a runtime value.
type ValueType int

const (
	VAL_NUMBER ValueType = iota
	VAL_STRING
	VAL_BOOL
	VAL_NULL
	VAL_ARRAY
	VAL_OBJECT
	VAL_FUNCTION
	VAL_BUILTIN
	VAL_RETURN    // wrapper for return values
	VAL_BREAK     // signal for break
	VAL_CONTINUE  // signal for continue
	VAL_FUTURE    // concurrent future (from spawn)
)

// Value represents a runtime value in Pengu.
type Value struct {
	Type     ValueType
	Number   float64
	Str      string
	Bool     bool
	Array    []*Value
	Object   *OrderedMap
	Func     *FunctionValue
	Builtin  BuiltinFunc
	Future   *Future
	IsInt    bool // whether the number is an integer
}

// FunctionValue holds the data for a user-defined function.
type FunctionValue struct {
	Name   string
	Params []string
	Body   []ast.Node
	Env    *Environment // closure environment
}

// BuiltinFunc is the signature for built-in functions.
type BuiltinFunc func(args []*Value) (*Value, error)

// Future holds the state for a spawned concurrent task.
type Future struct {
	ch     chan struct{} // closed when the result is ready
	result *Value
	err    error
	once   sync.Once
}

// Resolve is called by the goroutine to deliver the result.
func (f *Future) Resolve(val *Value, err error) {
	f.once.Do(func() {
		f.result = val
		f.err = err
		close(f.ch)
	})
}

// Await blocks until the future is resolved and returns the result.
func (f *Future) Await() (*Value, error) {
	<-f.ch
	return f.result, f.err
}

// OrderedMap preserves insertion order for object keys.
type OrderedMap struct {
	Keys   []string
	Values map[string]*Value
}

// NewOrderedMap creates a new empty ordered map.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:   make([]string, 0),
		Values: make(map[string]*Value),
	}
}

// Set adds or updates a key-value pair.
func (om *OrderedMap) Set(key string, val *Value) {
	if _, exists := om.Values[key]; !exists {
		om.Keys = append(om.Keys, key)
	}
	om.Values[key] = val
}

// Get retrieves a value by key.
func (om *OrderedMap) Get(key string) (*Value, bool) {
	v, ok := om.Values[key]
	return v, ok
}

// Len returns the number of entries.
func (om *OrderedMap) Len() int {
	return len(om.Keys)
}

// --- Value constructors ---

// NewNumber creates a number value.
func NewNumber(n float64, isInt bool) *Value {
	return &Value{Type: VAL_NUMBER, Number: n, IsInt: isInt}
}

// NewString creates a string value.
func NewString(s string) *Value {
	return &Value{Type: VAL_STRING, Str: s}
}

// NewBool creates a boolean value.
func NewBool(b bool) *Value {
	return &Value{Type: VAL_BOOL, Bool: b}
}

// NewNull creates a null value.
func NewNull() *Value {
	return &Value{Type: VAL_NULL}
}

// NewArray creates an array value.
func NewArray(elements []*Value) *Value {
	return &Value{Type: VAL_ARRAY, Array: elements}
}

// NewObject creates an object value.
func NewObject(om *OrderedMap) *Value {
	return &Value{Type: VAL_OBJECT, Object: om}
}

// NewFunction creates a function value.
func NewFunction(name string, params []string, body []ast.Node, env *Environment) *Value {
	return &Value{
		Type: VAL_FUNCTION,
		Func: &FunctionValue{
			Name:   name,
			Params: params,
			Body:   body,
			Env:    env,
		},
	}
}

// NewBuiltin creates a built-in function value.
func NewBuiltin(fn BuiltinFunc) *Value {
	return &Value{Type: VAL_BUILTIN, Builtin: fn}
}

// NewReturn creates a return signal wrapping a value.
func NewReturn(val *Value) *Value {
	return &Value{Type: VAL_RETURN, Array: []*Value{val}}
}

// NewBreak creates a break signal.
func NewBreak() *Value {
	return &Value{Type: VAL_BREAK}
}

// NewContinue creates a continue signal.
func NewContinue() *Value {
	return &Value{Type: VAL_CONTINUE}
}

// NewFuture creates a future value wrapping a channel-based result.
func NewFuture() (*Value, *Future) {
	f := &Future{
		ch: make(chan struct{}),
	}
	return &Value{Type: VAL_FUTURE, Future: f}, f
}

// Unwrap extracts the inner value from a return wrapper.
func (v *Value) Unwrap() *Value {
	if v.Type == VAL_RETURN && len(v.Array) > 0 {
		return v.Array[0]
	}
	return v
}

// IsTruthy returns whether the value is truthy.
func (v *Value) IsTruthy() bool {
	switch v.Type {
	case VAL_NULL:
		return false
	case VAL_BOOL:
		return v.Bool
	case VAL_NUMBER:
		return v.Number != 0
	case VAL_STRING:
		return v.Str != ""
	case VAL_ARRAY:
		return len(v.Array) > 0
	case VAL_OBJECT:
		return v.Object.Len() > 0
	default:
		return true
	}
}

// TypeName returns the Pengu type name of the value.
func (v *Value) TypeName() string {
	switch v.Type {
	case VAL_NUMBER:
		if v.IsInt {
			return "int"
		}
		return "float"
	case VAL_STRING:
		return "string"
	case VAL_BOOL:
		return "bool"
	case VAL_NULL:
		return "null"
	case VAL_ARRAY:
		return "array"
	case VAL_OBJECT:
		return "object"
	case VAL_FUNCTION:
		return "function"
	case VAL_BUILTIN:
		return "builtin"
	case VAL_FUTURE:
		return "future"
	default:
		return "unknown"
	}
}

// String returns the string representation of the value for display.
func (v *Value) String() string {
	switch v.Type {
	case VAL_NUMBER:
		if v.IsInt {
			return fmt.Sprintf("%d", int64(v.Number))
		}
		s := fmt.Sprintf("%g", v.Number)
		return s
	case VAL_STRING:
		return v.Str
	case VAL_BOOL:
		if v.Bool {
			return "true"
		}
		return "false"
	case VAL_NULL:
		return "null"
	case VAL_ARRAY:
		parts := make([]string, len(v.Array))
		for i, elem := range v.Array {
			if elem.Type == VAL_STRING {
				parts[i] = fmt.Sprintf("\"%s\"", elem.Str)
			} else {
				parts[i] = elem.String()
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case VAL_OBJECT:
		parts := make([]string, 0, v.Object.Len())
		for _, key := range v.Object.Keys {
			val := v.Object.Values[key]
			valStr := val.String()
			if val.Type == VAL_STRING {
				valStr = fmt.Sprintf("\"%s\"", val.Str)
			}
			parts = append(parts, fmt.Sprintf("\"%s\": %s", key, valStr))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case VAL_FUNCTION:
		return fmt.Sprintf("<fn %s>", v.Func.Name)
	case VAL_BUILTIN:
		return "<builtin>"
	case VAL_FUTURE:
		return "<future>"
	default:
		return "<unknown>"
	}
}
