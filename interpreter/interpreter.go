package interpreter

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/v4nsh0x/pengu/ast"
	"github.com/v4nsh0x/pengu/lexer"
	"github.com/v4nsh0x/pengu/parser"
	"github.com/v4nsh0x/pengu/runtime"
)

// Interpreter evaluates Pengu AST nodes.
type Interpreter struct {
	Global   *runtime.Environment
	basePath string
	imported map[string]bool
}

// New creates a new Interpreter with built-in functions registered.
func New() *Interpreter {
	interp := &Interpreter{
		Global:   runtime.NewEnvironment(nil),
		basePath: ".",
		imported: make(map[string]bool),
	}
	interp.registerBuiltins()
	return interp
}

// RunFile reads and executes a .pen file.
func (i *Interpreter) RunFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Error:\nCould not read file '%s'\n%s", filename, err)
	}
	absPath, _ := filepath.Abs(filename)
	i.basePath = filepath.Dir(absPath)
	return i.Run(string(data))
}

// Run executes Pengu source code.
func (i *Interpreter) Run(source string) error {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return err
	}
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return err
	}
	_, err = i.execProgram(program, i.Global)
	return err
}

func (i *Interpreter) execProgram(program *ast.Program, env *runtime.Environment) (*runtime.Value, error) {
	var result *runtime.Value
	for _, stmt := range program.Statements {
		val, err := i.exec(stmt, env)
		if err != nil {
			return nil, err
		}
		if val != nil && (val.Type == runtime.VAL_RETURN || val.Type == runtime.VAL_BREAK || val.Type == runtime.VAL_CONTINUE) {
			return val, nil
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) exec(node ast.Node, env *runtime.Environment) (*runtime.Value, error) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		return i.execVarDecl(n, env)
	case *ast.AssignmentExpression:
		return i.execAssignment(n, env)
	case *ast.FunctionDeclaration:
		return i.execFuncDecl(n, env)
	case *ast.ReturnStatement:
		return i.execReturn(n, env)
	case *ast.IfStatement:
		return i.execIf(n, env)
	case *ast.RepeatStatement:
		return i.execRepeat(n, env)
	case *ast.SayStatement:
		return i.execSay(n, env)
	case *ast.UseStatement:
		return i.execUse(n, env)
	case *ast.BreakStatement:
		return runtime.NewBreak(), nil
	case *ast.ContinueStatement:
		return runtime.NewContinue(), nil
	case *ast.TryCatchStatement:
		return i.execTryCatch(n, env)
	case *ast.BinaryExpression:
		return i.execBinary(n, env)
	case *ast.UnaryExpression:
		return i.execUnary(n, env)
	case *ast.CallExpression:
		return i.execCall(n, env)
	case *ast.IndexExpression:
		return i.execIndex(n, env)
	case *ast.MemberExpression:
		return i.execMember(n, env)
	case *ast.Identifier:
		val, err := env.Get(n.Name)
		if err != nil {
			return nil, fmt.Errorf("Runtime Error:\n%s\nLine %d", err, n.Line)
		}
		return val, nil
	case *ast.NumberLiteral:
		return runtime.NewNumber(n.Value, n.IsInt), nil
	case *ast.StringLiteral:
		return runtime.NewString(n.Value), nil
	case *ast.FStringLiteral:
		return i.execFString(n, env)
	case *ast.BooleanLiteral:
		return runtime.NewBool(n.Value), nil
	case *ast.NullLiteral:
		return runtime.NewNull(), nil
	case *ast.ArrayLiteral:
		return i.execArray(n, env)
	case *ast.ObjectLiteral:
		return i.execObject(n, env)
	default:
		return runtime.NewNull(), nil
	}
}

func (i *Interpreter) execVarDecl(n *ast.VariableDeclaration, env *runtime.Environment) (*runtime.Value, error) {
	val, err := i.exec(n.Value, env)
	if err != nil {
		return nil, err
	}
	env.Set(n.Name, val)
	return nil, nil
}

func (i *Interpreter) execAssignment(n *ast.AssignmentExpression, env *runtime.Environment) (*runtime.Value, error) {
	val, err := i.exec(n.Value, env)
	if err != nil {
		return nil, err
	}
	switch target := n.Target.(type) {
	case *ast.Identifier:
		if err := env.Update(target.Name, val); err != nil {
			return nil, fmt.Errorf("Runtime Error:\n%s\nLine %d", err, n.Line)
		}
	case *ast.IndexExpression:
		obj, err := i.exec(target.Object, env)
		if err != nil {
			return nil, err
		}
		idx, err := i.exec(target.Index, env)
		if err != nil {
			return nil, err
		}
		if obj.Type == runtime.VAL_ARRAY {
			index := int(idx.Number)
			if index < 0 || index >= len(obj.Array) {
				return nil, fmt.Errorf("Runtime Error:\nArray index %d out of bounds (length %d)\nLine %d", index, len(obj.Array), n.Line)
			}
			obj.Array[index] = val
		} else if obj.Type == runtime.VAL_OBJECT {
			obj.Object.Set(idx.Str, val)
		}
	case *ast.MemberExpression:
		obj, err := i.exec(target.Object, env)
		if err != nil {
			return nil, err
		}
		if obj.Type == runtime.VAL_OBJECT {
			obj.Object.Set(target.Property, val)
		}
	default:
		return nil, fmt.Errorf("Runtime Error:\nInvalid assignment target\nLine %d", n.Line)
	}
	return nil, nil
}

func (i *Interpreter) execFuncDecl(n *ast.FunctionDeclaration, env *runtime.Environment) (*runtime.Value, error) {
	fn := runtime.NewFunction(n.Name, n.Params, n.Body, env)
	if n.Name != "<anonymous>" {
		env.Set(n.Name, fn)
	}
	return fn, nil
}

func (i *Interpreter) execReturn(n *ast.ReturnStatement, env *runtime.Environment) (*runtime.Value, error) {
	val, err := i.exec(n.Value, env)
	if err != nil {
		return nil, err
	}
	return runtime.NewReturn(val), nil
}

func (i *Interpreter) execIf(n *ast.IfStatement, env *runtime.Environment) (*runtime.Value, error) {
	cond, err := i.exec(n.Condition, env)
	if err != nil {
		return nil, err
	}
	blockEnv := runtime.NewEnvironment(env)
	if cond.IsTruthy() {
		return i.execBlock(n.Body, blockEnv)
	} else if n.ElseBody != nil {
		return i.execBlock(n.ElseBody, blockEnv)
	}
	return nil, nil
}

func (i *Interpreter) execTryCatch(n *ast.TryCatchStatement, env *runtime.Environment) (*runtime.Value, error) {
	tryEnv := runtime.NewEnvironment(env)
	result, err := i.execBlock(n.TryBody, tryEnv)
	if err != nil {
		// An error occurred, execute catch block
		catchEnv := runtime.NewEnvironment(env)
		if n.CatchVar != "" {
			// Strip the "Runtime Error:\n" prefix for cleaner error messages in catch if we want,
			// or just give the raw error string
			catchEnv.Set(n.CatchVar, runtime.NewString(err.Error()))
		}
		// Reset the error, handle it
		return i.execBlock(n.CatchBody, catchEnv)
	}
	// No error occurred
	return result, nil
}

func (i *Interpreter) execFString(n *ast.FStringLiteral, env *runtime.Environment) (*runtime.Value, error) {
	str := n.Value
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(str, -1)

	for _, match := range matches {
		fullMatch := match[0]
		exprStr := match[1]

		// Lex and parse the expression snippet
		l := lexer.New(exprStr)
		tokens, err := l.Tokenize()
		if err != nil {
			return nil, fmt.Errorf("FString Lex Error:\n%v\nLine %d", err, n.Line)
		}

		p := parser.New(tokens)
		exprAst, err := p.ParseExpressionSnippet()
		if err != nil {
			return nil, fmt.Errorf("FString Parse Error:\n%v\nLine %d", err, n.Line)
		}

		val, err := i.exec(exprAst, env)
		if err != nil {
			return nil, err
		}

		str = strings.Replace(str, fullMatch, val.String(), 1)
	}

	return runtime.NewString(str), nil
}

func (i *Interpreter) execRepeat(n *ast.RepeatStatement, env *runtime.Environment) (*runtime.Value, error) {
	if n.Iterator != "" && n.Collection != nil {
		return i.execForEach(n, env)
	}
	if n.Condition != nil {
		return i.execWhile(n, env)
	}
	if n.Count != nil {
		return i.execCountLoop(n, env)
	}
	return nil, fmt.Errorf("Runtime Error:\nInvalid repeat statement\nLine %d", n.Line)
}

func (i *Interpreter) execCountLoop(n *ast.RepeatStatement, env *runtime.Environment) (*runtime.Value, error) {
	countVal, err := i.exec(n.Count, env)
	if err != nil {
		return nil, err
	}
	if countVal.Type != runtime.VAL_NUMBER {
		return nil, fmt.Errorf("Runtime Error:\nRepeat count must be a number\nLine %d", n.Line)
	}
	count := int(countVal.Number)
	for idx := 0; idx < count; idx++ {
		blockEnv := runtime.NewEnvironment(env)
		result, err := i.execBlock(n.Body, blockEnv)
		if err != nil {
			return nil, err
		}
		if result != nil {
			if result.Type == runtime.VAL_BREAK {
				break
			}
			if result.Type == runtime.VAL_CONTINUE {
				continue
			}
			if result.Type == runtime.VAL_RETURN {
				return result, nil
			}
		}
	}
	return nil, nil
}

func (i *Interpreter) execForEach(n *ast.RepeatStatement, env *runtime.Environment) (*runtime.Value, error) {
	collection, err := i.exec(n.Collection, env)
	if err != nil {
		return nil, err
	}
	switch collection.Type {
	case runtime.VAL_ARRAY:
		for idx, elem := range collection.Array {
			blockEnv := runtime.NewEnvironment(env)
			blockEnv.Set(n.Iterator, elem)
			if n.ValueIterator != "" {
				blockEnv.Set(n.ValueIterator, runtime.NewNumber(float64(idx), true))
			}
			result, err := i.execBlock(n.Body, blockEnv)
			if err != nil {
				return nil, err
			}
			if result != nil {
				if result.Type == runtime.VAL_BREAK {
					break
				}
				if result.Type == runtime.VAL_CONTINUE {
					continue
				}
				if result.Type == runtime.VAL_RETURN {
					return result, nil
				}
			}
		}
	case runtime.VAL_OBJECT:
		for _, key := range collection.Object.Keys {
			blockEnv := runtime.NewEnvironment(env)
			blockEnv.Set(n.Iterator, runtime.NewString(key))
			if n.ValueIterator != "" {
				val, _ := collection.Object.Get(key)
				blockEnv.Set(n.ValueIterator, val)
			}
			result, err := i.execBlock(n.Body, blockEnv)
			if err != nil {
				return nil, err
			}
			if result != nil {
				if result.Type == runtime.VAL_BREAK {
					break
				}
				if result.Type == runtime.VAL_CONTINUE {
					continue
				}
				if result.Type == runtime.VAL_RETURN {
					return result, nil
				}
			}
		}
	default:
		return nil, fmt.Errorf("Runtime Error:\nCannot iterate over %s\nLine %d", collection.TypeName(), n.Line)
	}
	return nil, nil
}

func (i *Interpreter) execWhile(n *ast.RepeatStatement, env *runtime.Environment) (*runtime.Value, error) {
	for {
		cond, err := i.exec(n.Condition, env)
		if err != nil {
			return nil, err
		}
		if !cond.IsTruthy() {
			break
		}
		blockEnv := runtime.NewEnvironment(env)
		result, err := i.execBlock(n.Body, blockEnv)
		if err != nil {
			return nil, err
		}
		if result != nil {
			if result.Type == runtime.VAL_BREAK {
				break
			}
			if result.Type == runtime.VAL_CONTINUE {
				continue
			}
			if result.Type == runtime.VAL_RETURN {
				return result, nil
			}
		}
	}
	return nil, nil
}

func (i *Interpreter) execSay(n *ast.SayStatement, env *runtime.Environment) (*runtime.Value, error) {
	val, err := i.exec(n.Value, env)
	if err != nil {
		return nil, err
	}
	fmt.Println(val.String())
	return nil, nil
}

func (i *Interpreter) execUse(n *ast.UseStatement, env *runtime.Environment) (*runtime.Value, error) {
	// Check for built-in native modules
	nativeModules := map[string]func() *runtime.Value{
		"http":   createHttpModule,
		"json":   createJsonModule,
		"os":     createOsModule,
		"crypto": createCryptoModule,
		"net":    createNetModule,
		"regex":  createRegexModule,
	}

	if creator, ok := nativeModules[n.Module]; ok {
		key := "__builtin_" + n.Module
		if !i.imported[key] {
			i.imported[key] = true
			env.Set(n.Module, creator())
		}
		return nil, nil
	}

	moduleName := n.Module + ".pen"

	// Search paths in order:
	// 1. Same directory as the current script
	// 2. modules/ directory next to the current script
	// 3. modules/ directory next to the pengu executable
	searchPaths := []string{
		filepath.Join(i.basePath, moduleName),
		filepath.Join(i.basePath, "modules", moduleName),
	}

	// Find the pengu executable's directory for built-in modules
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		searchPaths = append(searchPaths, filepath.Join(execDir, "modules", moduleName))
	}

	// Try each search path
	var modulePath string
	var data []byte
	for _, path := range searchPaths {
		d, err := os.ReadFile(path)
		if err == nil {
			modulePath = path
			data = d
			break
		}
	}

	if modulePath == "" {
		return nil, fmt.Errorf("Runtime Error:\nCould not import module '%s'\nSearched in:\n  %s\nLine %d",
			n.Module, strings.Join(searchPaths, "\n  "), n.Line)
	}

	absPath, _ := filepath.Abs(modulePath)
	if i.imported[absPath] {
		return nil, nil
	}
	i.imported[absPath] = true

	l := lexer.New(string(data))
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("Error in module '%s':\n%s", n.Module, err)
	}
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("Error in module '%s':\n%s", n.Module, err)
	}
	moduleEnv := runtime.NewEnvironment(i.Global)
	ret, err := i.execProgram(program, moduleEnv)
	if err != nil {
		return nil, fmt.Errorf("Error in module '%s':\n%s", n.Module, err)
	}
	
	if ret != nil {
		ret = ret.Unwrap()
		if ret.Type != runtime.VAL_NULL {
			env.Set(n.Module, ret)
		}
	}
	return nil, nil
}

func (i *Interpreter) execBlock(stmts []ast.Node, env *runtime.Environment) (*runtime.Value, error) {
	var result *runtime.Value
	for _, stmt := range stmts {
		val, err := i.exec(stmt, env)
		if err != nil {
			return nil, err
		}
		if val != nil && (val.Type == runtime.VAL_RETURN || val.Type == runtime.VAL_BREAK || val.Type == runtime.VAL_CONTINUE) {
			return val, nil
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) execBinary(n *ast.BinaryExpression, env *runtime.Environment) (*runtime.Value, error) {
	left, err := i.exec(n.Left, env)
	if err != nil {
		return nil, err
	}
	// Short-circuit for logical operators
	if n.Operator == "&&" {
		if !left.IsTruthy() {
			return left, nil
		}
		return i.exec(n.Right, env)
	}
	if n.Operator == "||" {
		if left.IsTruthy() {
			return left, nil
		}
		return i.exec(n.Right, env)
	}
	right, err := i.exec(n.Right, env)
	if err != nil {
		return nil, err
	}

	// String concatenation
	if n.Operator == "+" && (left.Type == runtime.VAL_STRING || right.Type == runtime.VAL_STRING) {
		return runtime.NewString(left.String() + right.String()), nil
	}

	// Equality
	if n.Operator == "==" {
		return runtime.NewBool(valuesEqual(left, right)), nil
	}
	if n.Operator == "!=" {
		return runtime.NewBool(!valuesEqual(left, right)), nil
	}

	// Numeric operations
	if left.Type == runtime.VAL_NUMBER && right.Type == runtime.VAL_NUMBER {
		return i.execNumericBinary(n.Operator, left, right, n.Line)
	}

	// String comparisons
	if left.Type == runtime.VAL_STRING && right.Type == runtime.VAL_STRING {
		switch n.Operator {
		case "<":
			return runtime.NewBool(left.Str < right.Str), nil
		case ">":
			return runtime.NewBool(left.Str > right.Str), nil
		case "<=":
			return runtime.NewBool(left.Str <= right.Str), nil
		case ">=":
			return runtime.NewBool(left.Str >= right.Str), nil
		}
	}

	return nil, fmt.Errorf("Runtime Error:\nCannot use '%s' with %s and %s\nLine %d", n.Operator, left.TypeName(), right.TypeName(), n.Line)
}

func (i *Interpreter) execNumericBinary(op string, left, right *runtime.Value, line int) (*runtime.Value, error) {
	a, b := left.Number, right.Number
	isInt := left.IsInt && right.IsInt
	switch op {
	case "+":
		return runtime.NewNumber(a+b, isInt), nil
	case "-":
		return runtime.NewNumber(a-b, isInt), nil
	case "*":
		return runtime.NewNumber(a*b, isInt), nil
	case "/":
		if b == 0 {
			return nil, fmt.Errorf("Runtime Error:\nDivision by zero\nLine %d", line)
		}
		result := a / b
		if isInt {
			return runtime.NewNumber(math.Trunc(result), true), nil
		}
		return runtime.NewNumber(result, false), nil
	case "%":
		if b == 0 {
			return nil, fmt.Errorf("Runtime Error:\nModulo by zero\nLine %d", line)
		}
		return runtime.NewNumber(math.Mod(a, b), isInt), nil
	case "<":
		return runtime.NewBool(a < b), nil
	case ">":
		return runtime.NewBool(a > b), nil
	case "<=":
		return runtime.NewBool(a <= b), nil
	case ">=":
		return runtime.NewBool(a >= b), nil
	default:
		return nil, fmt.Errorf("Runtime Error:\nUnknown operator '%s'\nLine %d", op, line)
	}
}

func (i *Interpreter) execUnary(n *ast.UnaryExpression, env *runtime.Environment) (*runtime.Value, error) {
	operand, err := i.exec(n.Operand, env)
	if err != nil {
		return nil, err
	}
	switch n.Operator {
	case "-":
		if operand.Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("Runtime Error:\nCannot negate %s\nLine %d", operand.TypeName(), n.Line)
		}
		return runtime.NewNumber(-operand.Number, operand.IsInt), nil
	case "!":
		return runtime.NewBool(!operand.IsTruthy()), nil
	default:
		return nil, fmt.Errorf("Runtime Error:\nUnknown unary operator '%s'\nLine %d", n.Operator, n.Line)
	}
}

func (i *Interpreter) execCall(n *ast.CallExpression, env *runtime.Environment) (*runtime.Value, error) {
	callee, err := i.exec(n.Callee, env)
	if err != nil {
		return nil, err
	}
	args := make([]*runtime.Value, len(n.Arguments))
	for idx, argNode := range n.Arguments {
		val, err := i.exec(argNode, env)
		if err != nil {
			return nil, err
		}
		args[idx] = val
	}
	switch callee.Type {
	case runtime.VAL_FUNCTION:
		return i.callFunction(callee.Func, args, n.Line)
	case runtime.VAL_BUILTIN:
		result, err := callee.Builtin(args)
		if err != nil {
			return nil, fmt.Errorf("Runtime Error:\n%s\nLine %d", err, n.Line)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Runtime Error:\nCannot call %s — it is not a function\nLine %d", callee.TypeName(), n.Line)
	}
}

func (i *Interpreter) callFunction(fn *runtime.FunctionValue, args []*runtime.Value, line int) (*runtime.Value, error) {
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("Runtime Error:\nFunction '%s' expects %d arguments but got %d\nLine %d",
			fn.Name, len(fn.Params), len(args), line)
	}
	funcEnv := runtime.NewEnvironment(fn.Env)
	for idx, param := range fn.Params {
		funcEnv.Set(param, args[idx])
	}
	result, err := i.execBlock(fn.Body, funcEnv)
	if err != nil {
		return nil, err
	}
	if result != nil && result.Type == runtime.VAL_RETURN {
		return result.Unwrap(), nil
	}
	if result != nil {
		return result, nil
	}
	return runtime.NewNull(), nil
}

func (i *Interpreter) execIndex(n *ast.IndexExpression, env *runtime.Environment) (*runtime.Value, error) {
	obj, err := i.exec(n.Object, env)
	if err != nil {
		return nil, err
	}
	idx, err := i.exec(n.Index, env)
	if err != nil {
		return nil, err
	}
	switch obj.Type {
	case runtime.VAL_ARRAY:
		if idx.Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("Runtime Error:\nArray index must be a number\nLine %d", n.Line)
		}
		index := int(idx.Number)
		if index < 0 || index >= len(obj.Array) {
			return nil, fmt.Errorf("Runtime Error:\nArray index %d out of bounds (length %d)\nLine %d", index, len(obj.Array), n.Line)
		}
		return obj.Array[index], nil
	case runtime.VAL_OBJECT:
		key := idx.Str
		if idx.Type != runtime.VAL_STRING {
			key = idx.String()
		}
		val, ok := obj.Object.Get(key)
		if !ok {
			return runtime.NewNull(), nil
		}
		return val, nil
	case runtime.VAL_STRING:
		if idx.Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("Runtime Error:\nString index must be a number\nLine %d", n.Line)
		}
		index := int(idx.Number)
		runes := []rune(obj.Str)
		if index < 0 || index >= len(runes) {
			return nil, fmt.Errorf("Runtime Error:\nString index %d out of bounds (length %d)\nLine %d", index, len(runes), n.Line)
		}
		return runtime.NewString(string(runes[index])), nil
	default:
		return nil, fmt.Errorf("Runtime Error:\nCannot index into %s\nLine %d", obj.TypeName(), n.Line)
	}
}

func (i *Interpreter) execMember(n *ast.MemberExpression, env *runtime.Environment) (*runtime.Value, error) {
	obj, err := i.exec(n.Object, env)
	if err != nil {
		return nil, err
	}
	if obj.Type == runtime.VAL_OBJECT {
		val, ok := obj.Object.Get(n.Property)
		if !ok {
			return runtime.NewNull(), nil
		}
		return val, nil
	}
	// Array/string built-in properties
	if obj.Type == runtime.VAL_ARRAY && n.Property == "length" {
		return runtime.NewNumber(float64(len(obj.Array)), true), nil
	}
	if obj.Type == runtime.VAL_STRING && n.Property == "length" {
		return runtime.NewNumber(float64(len([]rune(obj.Str))), true), nil
	}
	// Array methods
	if obj.Type == runtime.VAL_ARRAY {
		switch n.Property {
		case "push":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				for _, a := range args {
					obj.Array = append(obj.Array, a)
				}
				return runtime.NewNumber(float64(len(obj.Array)), true), nil
			}), nil
		case "pop":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(obj.Array) == 0 {
					return runtime.NewNull(), nil
				}
				last := obj.Array[len(obj.Array)-1]
				obj.Array = obj.Array[:len(obj.Array)-1]
				return last, nil
			}), nil
		case "map":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("map() expects 1 callback argument")
				}
				callback := args[0]
				resElems := make([]*runtime.Value, len(obj.Array))
				for idx, elem := range obj.Array {
					var cbArgs []*runtime.Value
					if callback.Type == runtime.VAL_FUNCTION {
						if len(callback.Func.Params) >= 2 {
							cbArgs = []*runtime.Value{elem, runtime.NewNumber(float64(idx), true)}
						} else {
							cbArgs = []*runtime.Value{elem}
						}
					} else {
						cbArgs = []*runtime.Value{elem}
					}

					val, err := i.invokeCallback(callback, cbArgs, n.Line)
					if err != nil {
						return nil, err
					}
					resElems[idx] = val
				}
				return runtime.NewArray(resElems), nil
			}), nil
		case "filter":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("filter() expects 1 callback argument")
				}
				callback := args[0]
				resElems := make([]*runtime.Value, 0)
				for idx, elem := range obj.Array {
					var cbArgs []*runtime.Value
					if callback.Type == runtime.VAL_FUNCTION {
						if len(callback.Func.Params) >= 2 {
							cbArgs = []*runtime.Value{elem, runtime.NewNumber(float64(idx), true)}
						} else {
							cbArgs = []*runtime.Value{elem}
						}
					} else {
						cbArgs = []*runtime.Value{elem}
					}

					val, err := i.invokeCallback(callback, cbArgs, n.Line)
					if err != nil {
						return nil, err
					}
					if val.IsTruthy() {
						resElems = append(resElems, elem)
					}
				}
				return runtime.NewArray(resElems), nil
			}), nil
		case "reduce":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("reduce() expects 1 callback and an optional initial value")
				}
				callback := args[0]
				var acc *runtime.Value
				startIdx := 0

				if len(args) == 2 {
					acc = args[1]
				} else {
					if len(obj.Array) == 0 {
						return nil, fmt.Errorf("reduce of empty array with no initial value")
					}
					acc = obj.Array[0]
					startIdx = 1
				}

				for idx := startIdx; idx < len(obj.Array); idx++ {
					elem := obj.Array[idx]
					var cbArgs []*runtime.Value
					if callback.Type == runtime.VAL_FUNCTION {
						if len(callback.Func.Params) >= 3 {
							cbArgs = []*runtime.Value{acc, elem, runtime.NewNumber(float64(idx), true)}
						} else {
							cbArgs = []*runtime.Value{acc, elem}
						}
					} else {
						cbArgs = []*runtime.Value{acc, elem}
					}

					val, err := i.invokeCallback(callback, cbArgs, n.Line)
					if err != nil {
						return nil, err
					}
					acc = val
				}
				return acc, nil
			}), nil
		case "find":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("find() expects 1 callback argument")
				}
				callback := args[0]
				for idx, elem := range obj.Array {
					var cbArgs []*runtime.Value
					if callback.Type == runtime.VAL_FUNCTION {
						if len(callback.Func.Params) >= 2 {
							cbArgs = []*runtime.Value{elem, runtime.NewNumber(float64(idx), true)}
						} else {
							cbArgs = []*runtime.Value{elem}
						}
					} else {
						cbArgs = []*runtime.Value{elem}
					}

					val, err := i.invokeCallback(callback, cbArgs, n.Line)
					if err != nil {
						return nil, err
					}
					if val.IsTruthy() {
						return elem, nil
					}
				}
				return runtime.NewNull(), nil
			}), nil
		case "includes":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("includes() expects 1 argument")
				}
				target := args[0]
				for _, elem := range obj.Array {
					if valuesEqual(elem, target) {
						return runtime.NewBool(true), nil
					}
				}
				return runtime.NewBool(false), nil
			}), nil
		case "reverse":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				resElems := make([]*runtime.Value, len(obj.Array))
				for idx, elem := range obj.Array {
					resElems[len(obj.Array)-1-idx] = elem
				}
				return runtime.NewArray(resElems), nil
			}), nil
		case "flat":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				resElems := make([]*runtime.Value, 0)
				for _, elem := range obj.Array {
					if elem.Type == runtime.VAL_ARRAY {
						resElems = append(resElems, elem.Array...)
					} else {
						resElems = append(resElems, elem)
					}
				}
				return runtime.NewArray(resElems), nil
			}), nil
		case "join":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				sep := ","
				if len(args) > 0 && args[0].Type == runtime.VAL_STRING {
					sep = args[0].Str
				}
				parts := make([]string, len(obj.Array))
				for idx, elem := range obj.Array {
					parts[idx] = elem.String()
				}
				return runtime.NewString(strings.Join(parts, sep)), nil
			}), nil
		}
	}
	// String methods
	if obj.Type == runtime.VAL_STRING {
		switch n.Property {
		case "upper":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				return runtime.NewString(strings.ToUpper(obj.Str)), nil
			}), nil
		case "lower":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				return runtime.NewString(strings.ToLower(obj.Str)), nil
			}), nil
		case "split":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				sep := " "
				if len(args) > 0 && args[0].Type == runtime.VAL_STRING {
					sep = args[0].Str
				}
				parts := strings.Split(obj.Str, sep)
				arr := make([]*runtime.Value, len(parts))
				for j, pt := range parts {
					arr[j] = runtime.NewString(pt)
				}
				return runtime.NewArray(arr), nil
			}), nil
		case "contains":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				if len(args) == 0 {
					return runtime.NewBool(false), nil
				}
				return runtime.NewBool(strings.Contains(obj.Str, args[0].String())), nil
			}), nil
		case "trim":
			return runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
				return runtime.NewString(strings.TrimSpace(obj.Str)), nil
			}), nil
		}
	}
	return nil, fmt.Errorf("Runtime Error:\nCannot access property '%s' on %s\nLine %d", n.Property, obj.TypeName(), n.Line)
}

func (i *Interpreter) execArray(n *ast.ArrayLiteral, env *runtime.Environment) (*runtime.Value, error) {
	elements := make([]*runtime.Value, len(n.Elements))
	for idx, elem := range n.Elements {
		val, err := i.exec(elem, env)
		if err != nil {
			return nil, err
		}
		elements[idx] = val
	}
	return runtime.NewArray(elements), nil
}

func (i *Interpreter) execObject(n *ast.ObjectLiteral, env *runtime.Environment) (*runtime.Value, error) {
	om := runtime.NewOrderedMap()
	for idx, keyNode := range n.Keys {
		keyVal, err := i.exec(keyNode, env)
		if err != nil {
			return nil, err
		}
		valVal, err := i.exec(n.Values[idx], env)
		if err != nil {
			return nil, err
		}
		om.Set(keyVal.Str, valVal)
	}
	return runtime.NewObject(om), nil
}

func valuesEqual(a, b *runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case runtime.VAL_NUMBER:
		return a.Number == b.Number
	case runtime.VAL_STRING:
		return a.Str == b.Str
	case runtime.VAL_BOOL:
		return a.Bool == b.Bool
	case runtime.VAL_NULL:
		return true
	default:
		return a == b // reference equality for arrays/objects
	}
}

func (i *Interpreter) invokeCallback(callback *runtime.Value, args []*runtime.Value, line int) (*runtime.Value, error) {
	switch callback.Type {
	case runtime.VAL_FUNCTION:
		return i.callFunction(callback.Func, args, line)
	case runtime.VAL_BUILTIN:
		return callback.Builtin(args)
	default:
		return nil, fmt.Errorf("Runtime Error:\nCallback must be a function, got %s\nLine %d", callback.TypeName(), line)
	}
}
