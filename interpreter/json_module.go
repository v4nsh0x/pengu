package interpreter

import (
	"encoding/json"
	"fmt"

	"github.com/v4nsh0x/pengu/runtime"
)

func createJsonModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	// json.parse(str) - Parse a JSON string into a Pengu value
	om.Set("parse", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("json.parse() expects 1 string argument")
		}
		var result interface{}
		err := json.Unmarshal([]byte(args[0].Str), &result)
		if err != nil {
			return nil, fmt.Errorf("json.parse() failed: %v", err)
		}
		return interfaceToValue(result), nil
	}))

	// json.stringify(value) - Convert a Pengu value to a JSON string
	om.Set("stringify", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.stringify() expects 1 argument")
		}
		data, err := json.Marshal(valueToInterface(args[0]))
		if err != nil {
			return nil, fmt.Errorf("json.stringify() failed: %v", err)
		}
		return runtime.NewString(string(data)), nil
	}))

	// json.pretty(value) - Convert a Pengu value to a pretty-printed JSON string
	om.Set("pretty", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.pretty() expects 1 argument")
		}
		data, err := json.MarshalIndent(valueToInterface(args[0]), "", "  ")
		if err != nil {
			return nil, fmt.Errorf("json.pretty() failed: %v", err)
		}
		return runtime.NewString(string(data)), nil
	}))

	return runtime.NewObject(om)
}
