package interpreter

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/v4nsh0x/pengu/runtime"
)

// registerBuiltins adds all built-in functions to the global environment.
func (i *Interpreter) registerBuiltins() {
	// say - also registered as a function for say() call syntax
	i.Global.Set("say", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		parts := make([]string, len(args))
		for idx, a := range args {
			parts[idx] = a.String()
		}
		fmt.Println(strings.Join(parts, " "))
		return runtime.NewNull(), nil
	}))

	// len
	i.Global.Set("len", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("len() expects 1 argument, got %d", len(args))
		}
		switch args[0].Type {
		case runtime.VAL_STRING:
			return runtime.NewNumber(float64(len([]rune(args[0].Str))), true), nil
		case runtime.VAL_ARRAY:
			return runtime.NewNumber(float64(len(args[0].Array)), true), nil
		case runtime.VAL_OBJECT:
			return runtime.NewNumber(float64(args[0].Object.Len()), true), nil
		default:
			return nil, fmt.Errorf("len() cannot be used on %s", args[0].TypeName())
		}
	}))

	// type
	i.Global.Set("type", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("type() expects 1 argument, got %d", len(args))
		}
		return runtime.NewString(args[0].TypeName()), nil
	}))

	// ask
	i.Global.Set("ask", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) > 0 {
			fmt.Print(args[0].String())
		}
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			return runtime.NewString(""), nil
		}
		text = strings.TrimRight(text, "\r\n")
		return runtime.NewString(text), nil
	}))

	// toInt
	i.Global.Set("toInt", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("toInt() expects 1 argument, got %d", len(args))
		}
		switch args[0].Type {
		case runtime.VAL_NUMBER:
			return runtime.NewNumber(math.Trunc(args[0].Number), true), nil
		case runtime.VAL_STRING:
			n, err := strconv.ParseFloat(args[0].Str, 64)
			if err != nil {
				return nil, fmt.Errorf("toInt() cannot convert '%s' to a number", args[0].Str)
			}
			return runtime.NewNumber(math.Trunc(n), true), nil
		case runtime.VAL_BOOL:
			if args[0].Bool {
				return runtime.NewNumber(1, true), nil
			}
			return runtime.NewNumber(0, true), nil
		default:
			return nil, fmt.Errorf("toInt() cannot convert %s to a number", args[0].TypeName())
		}
	}))

	// toFloat
	i.Global.Set("toFloat", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("toFloat() expects 1 argument, got %d", len(args))
		}
		switch args[0].Type {
		case runtime.VAL_NUMBER:
			return runtime.NewNumber(args[0].Number, false), nil
		case runtime.VAL_STRING:
			n, err := strconv.ParseFloat(args[0].Str, 64)
			if err != nil {
				return nil, fmt.Errorf("toFloat() cannot convert '%s' to a number", args[0].Str)
			}
			return runtime.NewNumber(n, false), nil
		default:
			return nil, fmt.Errorf("toFloat() cannot convert %s to a number", args[0].TypeName())
		}
	}))

	// toString
	i.Global.Set("toString", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("toString() expects 1 argument, got %d", len(args))
		}
		return runtime.NewString(args[0].String()), nil
	}))

	// append - add elements to an array
	i.Global.Set("append", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("append() expects at least 2 arguments")
		}
		if args[0].Type != runtime.VAL_ARRAY {
			return nil, fmt.Errorf("append() first argument must be an array")
		}
		newArr := make([]*runtime.Value, len(args[0].Array))
		copy(newArr, args[0].Array)
		newArr = append(newArr, args[1:]...)
		return runtime.NewArray(newArr), nil
	}))

	// keys - get object keys
	i.Global.Set("keys", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_OBJECT {
			return nil, fmt.Errorf("keys() expects 1 object argument")
		}
		arr := make([]*runtime.Value, len(args[0].Object.Keys))
		for idx, k := range args[0].Object.Keys {
			arr[idx] = runtime.NewString(k)
		}
		return runtime.NewArray(arr), nil
	}))

	// values - get object values
	i.Global.Set("values", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_OBJECT {
			return nil, fmt.Errorf("values() expects 1 object argument")
		}
		arr := make([]*runtime.Value, 0, len(args[0].Object.Keys))
		for _, k := range args[0].Object.Keys {
			arr = append(arr, args[0].Object.Values[k])
		}
		return runtime.NewArray(arr), nil
	}))

	// range - generate number array
	i.Global.Set("range", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 1 || len(args) > 3 {
			return nil, fmt.Errorf("range() expects 1-3 arguments")
		}
		var start, end, step float64
		switch len(args) {
		case 1:
			start, end, step = 0, args[0].Number, 1
		case 2:
			start, end, step = args[0].Number, args[1].Number, 1
		case 3:
			start, end, step = args[0].Number, args[1].Number, args[2].Number
		}
		if step == 0 {
			return nil, fmt.Errorf("range() step cannot be zero")
		}
		arr := make([]*runtime.Value, 0)
		if step > 0 {
			for v := start; v < end; v += step {
				arr = append(arr, runtime.NewNumber(v, true))
			}
		} else {
			for v := start; v > end; v += step {
				arr = append(arr, runtime.NewNumber(v, true))
			}
		}
		return runtime.NewArray(arr), nil
	}))

	// random - random number between 0 and 1, or 0 and max, or min and max
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	i.Global.Set("random", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) == 0 {
			return runtime.NewNumber(rng.Float64(), false), nil
		}
		if len(args) == 1 {
			max := int(args[0].Number)
			return runtime.NewNumber(float64(rng.Intn(max)), true), nil
		}
		if len(args) == 2 {
			min := int(args[0].Number)
			max := int(args[1].Number)
			return runtime.NewNumber(float64(min+rng.Intn(max-min)), true), nil
		}
		return nil, fmt.Errorf("random() expects 0-2 arguments")
	}))

	// floor, ceil, abs, sqrt, pow - math functions
	i.Global.Set("floor", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("floor() expects 1 number argument")
		}
		return runtime.NewNumber(math.Floor(args[0].Number), true), nil
	}))
	i.Global.Set("ceil", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("ceil() expects 1 number argument")
		}
		return runtime.NewNumber(math.Ceil(args[0].Number), true), nil
	}))
	i.Global.Set("abs", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("abs() expects 1 number argument")
		}
		return runtime.NewNumber(math.Abs(args[0].Number), args[0].IsInt), nil
	}))
	i.Global.Set("sqrt", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("sqrt() expects 1 number argument")
		}
		return runtime.NewNumber(math.Sqrt(args[0].Number), false), nil
	}))
	i.Global.Set("pow", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("pow() expects 2 arguments")
		}
		if args[0].Type != runtime.VAL_NUMBER || args[1].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("pow() expects number arguments")
		}
		result := math.Pow(args[0].Number, args[1].Number)
		isInt := args[0].IsInt && args[1].IsInt && args[1].Number >= 0
		return runtime.NewNumber(result, isInt), nil
	}))

	// await_all - waits for an array of futures to all resolve
	i.Global.Set("await_all", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_ARRAY {
			return nil, fmt.Errorf("await_all() expects an array of futures")
		}
		futures := args[0].Array
		results := make([]*runtime.Value, len(futures))

		// Await all futures concurrently using goroutines
		type indexedResult struct {
			idx int
			val *runtime.Value
			err error
		}
		ch := make(chan indexedResult, len(futures))
		for idx, f := range futures {
			if f.Type != runtime.VAL_FUTURE {
				return nil, fmt.Errorf("await_all() expects all elements to be futures, got %s at index %d", f.TypeName(), idx)
			}
			go func(i int, fut *runtime.Value) {
				val, err := fut.Future.Await()
				ch <- indexedResult{i, val, err}
			}(idx, f)
		}

		for range futures {
			res := <-ch
			if res.err != nil {
				return nil, fmt.Errorf("Runtime Error (in spawned task #%d):\n%s", res.idx, res.err)
			}
			results[res.idx] = res.val
		}

		return runtime.NewArray(results), nil
	}))
}
