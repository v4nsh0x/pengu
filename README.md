# 🐧 Pengu

**A fun, fast, and friendly programming language.**

Pengu is a lightweight, expressive scripting language built in Go. It's designed to feel simple like Python, lightweight like Lua, and clean like Go — while being playful and beginner-friendly.

```pen
store name = "world"

fn greet(who) {
    say "Hello, " + who + "! 🐧"
}

greet(name)
```

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Language Guide](#language-guide)
  - [Variables](#variables)
  - [Data Types](#data-types)
  - [Output](#output)
  - [Input](#input)
  - [Operators](#operators)
  - [Conditions](#conditions)
  - [Loops](#loops)
  - [Functions](#functions)
  - [Arrays](#arrays)
  - [Objects](#objects)
  - [Imports](#imports)
  - [Comments](#comments)
- [Built-in Functions](#built-in-functions)
- [String Methods](#string-methods)
- [Array Methods](#array-methods)
- [Error Handling](#error-handling)
- [Architecture](#architecture)
- [Examples](#examples)

---

## Installation

### Build from Source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/vansh/pengu.git
cd pengu
go build -o pengu .
```

The `pengu` binary is now ready to use.

---

## Quick Start

Create a file called `hello.pen`:

```pen
say "Hello, Pengu! 🐧"
```

Run it:

```bash
pengu hello.pen
```

Output:

```
Hello, Pengu! 🐧
```

---

## CLI Usage

### Run a Script

```bash
pengu <file.pen>
pengu run <file.pen>
```

Both forms are equivalent.

### Interactive REPL

```bash
pengu repl
```

```
🐧 > store x = 42
🐧 > say x
42
🐧 > exit
Bye! 🐧
```

Type `exit` or `quit` to leave the REPL.

### Compile to Executable

```bash
pengu build <file.pen> -o <output>
pengu <file.pen> -o <output>
```

Examples:

```bash
pengu build game.pen -o game
pengu server.pen -o api
```

The generated binary:
- Runs independently (no Pengu installation needed)
- Bundles the Pengu runtime
- Automatically gets `.exe` on Windows, no extension on Linux/macOS

### Version & Help

```bash
pengu version
pengu help
```

---

## Language Guide

### Variables

Declare variables with `store`. Variables are dynamically typed.

```pen
store name = "vansh"
store age = 18
store pi = 3.14
store online = true
store nothing = null
```

Reassign existing variables without `store`:

```pen
store x = 10
x = 20
```

> **Note:** You must use `store` to declare a variable before reassigning it. Assigning to an undeclared variable produces an error.

---

### Data Types

Pengu has 6 data types:

| Type | Example | Type Name |
|------|---------|-----------|
| Integer | `42` | `int` |
| Float | `3.14` | `float` |
| String | `"hello"` | `string` |
| Boolean | `true`, `false` | `bool` |
| Null | `null` | `null` |
| Array | `[1, 2, 3]` | `array` |
| Object | `{"key": "val"}` | `object` |
| Function | `fn(x) { return x }` | `function` |

Check a value's type with `type()`:

```pen
say type(42)        // "int"
say type(3.14)      // "float"
say type("hello")   // "string"
say type(true)      // "bool"
say type(null)      // "null"
say type([1,2])     // "array"
```

---

### Output

Use `say` to print to the console. Both statement and function forms work:

```pen
// Statement form (no parentheses)
say "hello world"
say 42
say name

// Function form (with parentheses)
say("hello world")
say(42)
say(name)
```

`say` can take multiple arguments when called as a function:

```pen
say("hello", "world", 42)
// Output: hello world 42
```

`say` always adds a newline at the end.

---

### Input

Read user input with `ask()`:

```pen
store name = ask("What is your name? ")
say "Hello, " + name + "!"
```

`ask()` takes an optional prompt string and returns the user's input as a string.

---

### Operators

#### Arithmetic

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `5 + 3` → `8` |
| `-` | Subtraction | `5 - 3` → `2` |
| `*` | Multiplication | `5 * 3` → `15` |
| `/` | Division | `10 / 3` → `3` (integer) |
| `%` | Modulo | `10 % 3` → `1` |

> **Integer Division:** When both operands are integers, `/` performs integer division (truncates). Use `toFloat()` for float division.

```pen
say 10 / 3          // 3   (integer division)
say toFloat(10) / 3 // 3.3333...
```

#### String Concatenation

The `+` operator concatenates strings. If either side is a string, the other is converted automatically:

```pen
say "hello " + "world"   // "hello world"
say "age: " + 18         // "age: 18"  (note: no toString needed with +)
```

#### Comparison

| Operator | Description |
|----------|-------------|
| `==` | Equal |
| `!=` | Not equal |
| `<` | Less than |
| `>` | Greater than |
| `<=` | Less than or equal |
| `>=` | Greater than or equal |

Works with numbers and strings:

```pen
say 5 > 3         // true
say "a" < "b"     // true
say 10 == 10      // true
say "hi" != "bye" // true
```

#### Logical

| Operator | Description |
|----------|-------------|
| `&&` | Logical AND (short-circuit) |
| `\|\|` | Logical OR (short-circuit) |
| `!` | Logical NOT |

```pen
say true && false  // false
say true || false  // true
say !true          // false
```

#### Unary

| Operator | Description | Example |
|----------|-------------|---------|
| `-` | Negation | `-5` |
| `!` | Logical NOT | `!true` → `false` |

#### Operator Precedence (high to low)

1. `!`, `-` (unary)
2. `*`, `/`, `%`
3. `+`, `-`
4. `<`, `>`, `<=`, `>=`
5. `==`, `!=`
6. `&&`
7. `||`

Use parentheses `()` to override precedence:

```pen
say (2 + 3) * 4  // 20
```

---

### Conditions

Use `when` for if-statements and `otherwise` for else.

#### Basic Condition

```pen
when age >= 18 {
    say "adult"
}
```

#### With Otherwise (else)

```pen
when score > 50 {
    say "pass"
} otherwise {
    say "fail"
}
```

#### Nested Conditions

```pen
when score >= 90 {
    say "A"
} otherwise {
    when score >= 80 {
        say "B"
    } otherwise {
        when score >= 70 {
            say "C"
        } otherwise {
            say "F"
        }
    }
}
```

#### Truthiness

Values are truthy/falsy as follows:

| Value | Truthy? |
|-------|---------|
| `false` | ❌ |
| `null` | ❌ |
| `0` | ❌ |
| `""` (empty string) | ❌ |
| `[]` (empty array) | ❌ |
| `{}` (empty object) | ❌ |
| Everything else | ✅ |

---

### Loops

Pengu uses `repeat` for all loop types.

#### Numeric Loop (repeat N times)

```pen
repeat 5 {
    say "hello"
}
```

#### For-Each Loop

Iterate over arrays:

```pen
store fruits = ["apple", "banana", "cherry"]
repeat fruit in fruits {
    say fruit
}
```

Iterate over object keys:

```pen
store user = {"name": "vansh", "age": 18}
repeat key in user {
    say key
}
// Output: name, age
```

#### Conditional Loop (while)

```pen
store x = 0
repeat x < 10 {
    say x
    x = x + 1
}
```

#### Break & Continue

`break` exits the loop, `continue` skips to the next iteration:

```pen
repeat i in range(10) {
    when i == 5 {
        break
    }
    when i % 2 == 0 {
        continue
    }
    say i
}
// Output: 1, 3
```

---

### Functions

Define functions with `fn`.

#### Basic Function

```pen
fn greet(name) {
    say "Hello, " + name + "!"
}

greet("Pengu")
```

#### Return Values

```pen
fn add(a, b) {
    return a + b
}

store result = add(2, 3)
say result  // 5
```

If no `return` is specified, the function returns `null`.

#### Recursion

```pen
fn factorial(n) {
    when n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

say factorial(10)  // 3628800
```

#### Anonymous Functions (Lambdas)

```pen
store double = fn(x) {
    return x * 2
}

say double(5)  // 10
```

#### Closures

Functions capture their enclosing scope:

```pen
fn makeCounter() {
    store count = 0
    return fn() {
        count = count + 1
        return count
    }
}

store counter = makeCounter()
say counter()  // 1
say counter()  // 2
say counter()  // 3
```

#### Higher-Order Functions

Functions can be passed as arguments:

```pen
fn apply(f, x) {
    return f(x)
}

fn square(n) {
    return n * n
}

say apply(square, 5)  // 25
```

---

### Arrays

Arrays are ordered lists that can hold any type.

#### Creating Arrays

```pen
store nums = [1, 2, 3, 4, 5]
store mixed = [1, "hello", true, null]
store nested = [[1, 2], [3, 4]]
store empty = []
```

#### Accessing Elements

Zero-indexed:

```pen
say nums[0]    // 1
say nums[2]    // 3
say nested[0]  // [1, 2]
say nested[0][1]  // 2
```

#### Modifying Elements

```pen
nums[0] = 99
say nums  // [99, 2, 3, 4, 5]
```

#### Array Length

```pen
say len(nums)       // 5
say nums.length     // 5 (property access)
```

#### Array Methods

```pen
// Push — add to end (mutates the array)
nums.push(6)
say nums  // [1, 2, 3, 4, 5, 6]

// Pop — remove from end (mutates the array)
store last = nums.pop()
say last  // 6

// Append — create a new array with added elements (does NOT mutate)
store newArr = append(nums, 7, 8)
say newArr  // [1, 2, 3, 4, 5, 7, 8]
```

#### Iterating Arrays

```pen
repeat item in nums {
    say item
}
```

---

### Objects

Objects are key-value maps with string keys. Syntax is JSON-like.

#### Creating Objects

```pen
store user = {
    "name": "vansh",
    "age": 18,
    "skills": ["go", "pengu"]
}
```

Keys can also be unquoted identifiers:

```pen
store config = {
    host: "localhost",
    port: 8080
}
```

#### Accessing Values

Bracket notation:

```pen
say user["name"]   // "vansh"
say user["age"]    // 18
```

Dot notation:

```pen
say user.name      // (only for object values, not built-in properties)
```

#### Modifying Values

```pen
user["age"] = 19
user["email"] = "vansh@example.com"  // add new key
```

#### Object Functions

```pen
// Get all keys as an array
say keys(user)    // ["name", "age", "skills"]

// Get all values as an array
say values(user)  // ["vansh", 18, ["go", "pengu"]]

// Get number of keys
say len(user)     // 3
```

#### Iterating Over Objects

```pen
repeat key in user {
    say key + ": " + toString(user[key])
}
```

#### Nested Objects

```pen
store data = {
    "user": {
        "name": "vansh",
        "address": {
            "city": "Delhi"
        }
    }
}

say data["user"]["address"]["city"]  // "Delhi"
```

---

### Imports

Use `use` to import other `.pen` files. The imported file's functions and variables become available in the current scope.

#### Creating a Module

`math.pen`:

```pen
fn square(x) {
    return x * x
}

fn cube(x) {
    return x * x * x
}
```

#### Importing

`main.pen`:

```pen
use math

say square(5)  // 25
say cube(3)    // 27
```

> **How it works:** `use math` looks for `math.pen` in the same directory as the current file and executes it in the current scope.

> **Note:** Each module is imported only once, even if `use` is called multiple times.

---

### Comments

#### Single-Line Comments

```pen
// This is a comment
store x = 5  // inline comment
```

#### Multi-Line Comments

```pen
/*
  This is a
  multi-line comment
*/
```

---

## Built-in Functions

### I/O

| Function | Description | Example |
|----------|-------------|---------|
| `say(args...)` | Print values to console with newline | `say("hello", "world")` |
| `ask(prompt?)` | Read a line from stdin | `store name = ask("Name: ")` |

### Type Functions

| Function | Description | Returns |
|----------|-------------|---------|
| `type(value)` | Get the type name of a value | `string` |
| `len(value)` | Get length of string, array, or object | `int` |

```pen
say type(42)       // "int"
say type(3.14)     // "float"
say type("hello")  // "string"
say type(true)     // "bool"
say type(null)     // "null"
say type([1,2])    // "array"
say type({})       // "object"

say len("hello")   // 5
say len([1,2,3])   // 3
say len({"a": 1})  // 1
```

### Conversion Functions

| Function | Description | Example |
|----------|-------------|---------|
| `toInt(value)` | Convert to integer | `toInt("42")` → `42` |
| `toFloat(value)` | Convert to float | `toFloat("3.14")` → `3.14` |
| `toString(value)` | Convert to string | `toString(42)` → `"42"` |

```pen
store num = toInt("42")
say num + 1  // 43

store pi = toFloat("3.14")
say pi * 2  // 6.28

store s = toString([1, 2, 3])
say s  // "[1, 2, 3]"
```

`toInt` truncates floats:

```pen
say toInt(3.9)  // 3
```

`toInt` and `toFloat` convert booleans:

```pen
say toInt(true)   // 1
say toInt(false)  // 0
```

### Array Functions

| Function | Description | Example |
|----------|-------------|---------|
| `append(array, items...)` | Create new array with items added | `append([1,2], 3, 4)` → `[1,2,3,4]` |
| `range(end)` | Generate `[0, 1, ..., end-1]` | `range(5)` → `[0,1,2,3,4]` |
| `range(start, end)` | Generate `[start, ..., end-1]` | `range(2, 5)` → `[2,3,4]` |
| `range(start, end, step)` | Generate with step | `range(0, 10, 2)` → `[0,2,4,6,8]` |

```pen
// append creates a NEW array (does not mutate the original)
store a = [1, 2]
store b = append(a, 3, 4)
say a  // [1, 2]
say b  // [1, 2, 3, 4]

// range with different argument counts
say range(5)        // [0, 1, 2, 3, 4]
say range(2, 6)     // [2, 3, 4, 5]
say range(0, 10, 3) // [0, 3, 6, 9]
say range(5, 0, -1) // [5, 4, 3, 2, 1]
```

### Object Functions

| Function | Description | Example |
|----------|-------------|---------|
| `keys(object)` | Get all keys as array | `keys({"a": 1})` → `["a"]` |
| `values(object)` | Get all values as array | `values({"a": 1})` → `[1]` |

### Math Functions

| Function | Description | Example |
|----------|-------------|---------|
| `floor(n)` | Round down | `floor(3.7)` → `3` |
| `ceil(n)` | Round up | `ceil(3.2)` → `4` |
| `abs(n)` | Absolute value | `abs(-42)` → `42` |
| `sqrt(n)` | Square root | `sqrt(16)` → `4` |
| `pow(base, exp)` | Exponentiation | `pow(2, 10)` → `1024` |
| `random()` | Random float 0.0–1.0 | `random()` → `0.7291...` |
| `random(max)` | Random int 0 to max-1 | `random(10)` → `7` |
| `random(min, max)` | Random int min to max-1 | `random(5, 10)` → `8` |

```pen
say floor(3.7)    // 3
say ceil(3.2)     // 4
say abs(-42)      // 42
say sqrt(16)      // 4
say pow(2, 10)    // 1024
say random()      // 0.7291... (varies)
say random(100)   // 42 (varies)
say random(1, 6)  // 3 (varies, like a dice roll)
```

---

## String Methods

Strings have built-in methods accessible via dot notation:

| Method | Description | Example |
|--------|-------------|---------|
| `.upper()` | Convert to uppercase | `"hello".upper()` → `"HELLO"` |
| `.lower()` | Convert to lowercase | `"HELLO".lower()` → `"hello"` |
| `.trim()` | Remove leading/trailing whitespace | `"  hi  ".trim()` → `"hi"` |
| `.split(sep?)` | Split into array | `"a,b,c".split(",")` → `["a","b","c"]` |
| `.contains(sub)` | Check if contains substring | `"hello".contains("ell")` → `true` |
| `.length` | Get string length (property) | `"hello".length` → `5` |

```pen
store msg = "Hello World"

say msg.upper()           // "HELLO WORLD"
say msg.lower()           // "hello world"
say msg.contains("World") // true
say msg.split(" ")        // ["Hello", "World"]
say msg.length            // 11

store padded = "  trim me  "
say padded.trim()         // "trim me"

// split with default separator (space)
say "a b c".split()       // ["a", "b", "c"]
```

### String Indexing

Access individual characters by index:

```pen
store s = "Pengu"
say s[0]  // "P"
say s[4]  // "u"
```

---

## Array Methods

Arrays have built-in methods accessible via dot notation:

| Method | Description | Returns |
|--------|-------------|---------|
| `.push(items...)` | Add items to end (mutates) | New length |
| `.pop()` | Remove & return last item (mutates) | Removed item |
| `.length` | Get array length (property) | Number |

```pen
store colors = ["red", "green"]

colors.push("blue")
say colors          // ["red", "green", "blue"]

store removed = colors.pop()
say removed         // "blue"
say colors          // ["red", "green"]

say colors.length   // 2
```

---

## Error Handling

Pengu provides clear, friendly error messages with line numbers.

### Syntax Errors

```
Syntax Error:
Expected '}' to close block
Line 12, Column 1
```

### Runtime Errors

```
Runtime Error:
'x' is not defined
Line 5
```

```
Runtime Error:
Array index 10 out of bounds (length 3)
Line 8
```

```
Runtime Error:
Division by zero
Line 3
```

```
Runtime Error:
Function 'greet' expects 1 arguments but got 0
Line 7
```

### Import Errors

```
Runtime Error:
Could not import module 'utils'
File not found: utils.pen
Line 1
```

---

## Architecture

Pengu is implemented as a tree-walking interpreter in Go.

```
Source Code (.pen)
       │
       ▼
   ┌────────┐
   │  Lexer  │  → Tokenizes source into tokens
   └────┬───┘
        │
        ▼
   ┌────────┐
   │ Parser  │  → Builds Abstract Syntax Tree
   └────┬───┘
        │
        ▼
   ┌─────────────┐
   │ Interpreter  │  → Walks the AST and evaluates
   └──────┬──────┘
          │
          ▼
      Output
```

### Project Structure

```
pengu/
├── ast/             # AST node definitions
│   └── ast.go
├── lexer/           # Tokenizer
│   ├── token.go
│   └── lexer.go
├── parser/          # Recursive descent parser
│   └── parser.go
├── interpreter/     # Tree-walking evaluator
│   ├── interpreter.go
│   └── builtins.go
├── runtime/         # Value types & environments
│   ├── values.go
│   └── environment.go
├── cli/             # CLI (run, repl, build)
│   └── cli.go
├── examples/        # Example .pen programs
├── main.go          # Entry point
└── go.mod
```

---

## Examples

### Hello World

```pen
say "Hello, World! 🐧"
```

### FizzBuzz

```pen
repeat i in range(1, 21) {
    when i % 15 == 0 {
        say "FizzBuzz"
    } otherwise {
        when i % 3 == 0 {
            say "Fizz"
        } otherwise {
            when i % 5 == 0 {
                say "Buzz"
            } otherwise {
                say i
            }
        }
    }
}
```

### Fibonacci

```pen
fn fib(n) {
    when n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

repeat i in range(10) {
    say fib(i)
}
```

### Todo List

```pen
store todos = []

fn addTodo(text) {
    todos = append(todos, {"text": text, "done": false})
}

fn showTodos() {
    repeat i in range(len(todos)) {
        store todo = todos[i]
        store status = "[ ]"
        when todo["done"] {
            status = "[x]"
        }
        say toString(i) + ". " + status + " " + todo["text"]
    }
}

addTodo("Learn Pengu")
addTodo("Build something cool")
addTodo("Have fun")

showTodos()
```

### Number Guessing Game

```pen
store secret = random(1, 101)
store guesses = 0

say "I'm thinking of a number between 1 and 100!"

repeat true {
    store guess = toInt(ask("Your guess: "))
    guesses = guesses + 1

    when guess == secret {
        say "🎉 Correct! You got it in " + toString(guesses) + " guesses!"
        break
    } otherwise {
        when guess < secret {
            say "Too low!"
        } otherwise {
            say "Too high!"
        }
    }
}
```

---

## License

MIT

---

*Built with ❤️ and Go. Happy coding! 🐧*
