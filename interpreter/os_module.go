package interpreter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/v4nsh0x/pengu/runtime"
)

func createOsModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	// os.exec(command) - Execute a system command and return stdout, stderr, exit_code
	om.Set("exec", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.exec() expects a string command")
		}

		var cmd *exec.Cmd
		if goruntime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", args[0].Str)
		} else {
			cmd = exec.Command("sh", "-c", args[0].Str)
		}

		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				return nil, fmt.Errorf("os.exec() failed: %v", err)
			}
		}

		result := runtime.NewOrderedMap()
		result.Set("stdout", runtime.NewString(strings.TrimRight(stdout.String(), "\r\n")))
		result.Set("stderr", runtime.NewString(strings.TrimRight(stderr.String(), "\r\n")))
		result.Set("exit_code", runtime.NewNumber(float64(exitCode), true))
		return runtime.NewObject(result), nil
	}))

	// os.env(name) - Get environment variable
	om.Set("env", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.env() expects a string argument")
		}
		val := os.Getenv(args[0].Str)
		if val == "" && len(args) >= 2 {
			return args[1], nil // return fallback
		}
		return runtime.NewString(val), nil
	}))

	// os.exists(path) - Check if a file or directory exists
	om.Set("exists", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.exists() expects a string path")
		}
		_, err := os.Stat(args[0].Str)
		return runtime.NewBool(!os.IsNotExist(err)), nil
	}))

	// os.read_file(path) - Read entire file contents as a string
	om.Set("read_file", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.read_file() expects a string path")
		}
		data, err := os.ReadFile(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("os.read_file() failed: %v", err)
		}
		return runtime.NewString(string(data)), nil
	}))

	// os.write_file(path, content) - Write content to a file (overwrites)
	om.Set("write_file", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.write_file() expects (path, content) as strings")
		}
		err := os.WriteFile(args[0].Str, []byte(args[1].Str), 0644)
		if err != nil {
			return nil, fmt.Errorf("os.write_file() failed: %v", err)
		}
		return runtime.NewBool(true), nil
	}))

	// os.append_file(path, content) - Append content to a file
	om.Set("append_file", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.append_file() expects (path, content) as strings")
		}
		f, err := os.OpenFile(args[0].Str, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("os.append_file() failed: %v", err)
		}
		defer f.Close()
		_, err = io.WriteString(f, args[1].Str)
		if err != nil {
			return nil, fmt.Errorf("os.append_file() write failed: %v", err)
		}
		return runtime.NewBool(true), nil
	}))

	// os.remove(path) - Delete a file
	om.Set("remove", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.remove() expects a string path")
		}
		err := os.Remove(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("os.remove() failed: %v", err)
		}
		return runtime.NewBool(true), nil
	}))

	// os.mkdir(path) - Create a directory (and parents)
	om.Set("mkdir", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.mkdir() expects a string path")
		}
		err := os.MkdirAll(args[0].Str, 0755)
		if err != nil {
			return nil, fmt.Errorf("os.mkdir() failed: %v", err)
		}
		return runtime.NewBool(true), nil
	}))

	// os.list_dir(path) - List directory contents
	om.Set("list_dir", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.list_dir() expects a string path")
		}
		entries, err := os.ReadDir(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("os.list_dir() failed: %v", err)
		}
		arr := make([]*runtime.Value, 0, len(entries))
		for _, entry := range entries {
			item := runtime.NewOrderedMap()
			item.Set("name", runtime.NewString(entry.Name()))
			item.Set("is_dir", runtime.NewBool(entry.IsDir()))
			info, err := entry.Info()
			if err == nil {
				item.Set("size", runtime.NewNumber(float64(info.Size()), true))
			}
			arr = append(arr, runtime.NewObject(item))
		}
		return runtime.NewArray(arr), nil
	}))

	// os.cwd() - Get current working directory
	om.Set("cwd", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		dir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("os.cwd() failed: %v", err)
		}
		return runtime.NewString(dir), nil
	}))

	// os.platform() - Get OS name
	om.Set("platform", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewString(goruntime.GOOS), nil
	}))

	// os.arch() - Get architecture
	om.Set("arch", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		return runtime.NewString(goruntime.GOARCH), nil
	}))

	// os.abs_path(path) - Get absolute path
	om.Set("abs_path", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("os.abs_path() expects a string path")
		}
		abs, err := filepath.Abs(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("os.abs_path() failed: %v", err)
		}
		return runtime.NewString(abs), nil
	}))

	// os.exit(code) - Exit the program
	om.Set("exit", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		code := 0
		if len(args) >= 1 && args[0].Type == runtime.VAL_NUMBER {
			code = int(args[0].Number)
		}
		os.Exit(code)
		return runtime.NewNull(), nil
	}))

	return runtime.NewObject(om)
}
