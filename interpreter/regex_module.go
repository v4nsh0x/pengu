package interpreter

import (
	"fmt"
	"regexp"

	"github.com/v4nsh0x/pengu/runtime"
)

func createRegexModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	// regex.match(pattern, text) -> boolean
	om.Set("match", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.match() expects (pattern, text)")
		}

		matched, err := regexp.MatchString(args[0].Str, args[1].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.match() error: %v", err)
		}

		return runtime.NewBool(matched), nil
	}))

	// regex.find(pattern, text) -> string or null
	om.Set("find", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.find() expects (pattern, text)")
		}

		re, err := regexp.Compile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.find() error: %v", err)
		}

		match := re.FindString(args[1].Str)
		if match == "" {
			return runtime.NewNull(), nil
		}
		return runtime.NewString(match), nil
	}))

	// regex.find_all(pattern, text) -> array of strings
	om.Set("find_all", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.find_all() expects (pattern, text)")
		}

		re, err := regexp.Compile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.find_all() error: %v", err)
		}

		matches := re.FindAllString(args[1].Str, -1)
		var elements []*runtime.Value
		for _, m := range matches {
			elements = append(elements, runtime.NewString(m))
		}

		return runtime.NewArray(elements), nil
	}))

	// regex.replace(pattern, replacement, text) -> string
	om.Set("replace", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 3 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING || args[2].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.replace() expects (pattern, replacement, text)")
		}

		re, err := regexp.Compile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.replace() error: %v", err)
		}

		result := re.ReplaceAllString(args[2].Str, args[1].Str)
		return runtime.NewString(result), nil
	}))

	// regex.split(pattern, text) -> array of strings
	om.Set("split", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.split() expects (pattern, text)")
		}

		re, err := regexp.Compile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.split() error: %v", err)
		}

		parts := re.Split(args[1].Str, -1)
		var elements []*runtime.Value
		for _, p := range parts {
			elements = append(elements, runtime.NewString(p))
		}

		return runtime.NewArray(elements), nil
	}))

	// regex.extract(pattern, text) -> array of submatches for the first match
	om.Set("extract", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("regex.extract() expects (pattern, text)")
		}

		re, err := regexp.Compile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("regex.extract() error: %v", err)
		}

		submatches := re.FindStringSubmatch(args[1].Str)
		if len(submatches) == 0 {
			return runtime.NewArray([]*runtime.Value{}), nil
		}

		var elements []*runtime.Value
		for _, m := range submatches {
			elements = append(elements, runtime.NewString(m))
		}

		return runtime.NewArray(elements), nil
	}))

	return runtime.NewObject(om)
}
