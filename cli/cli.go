package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/v4nsh0x/pengu/interpreter"
	"github.com/v4nsh0x/pengu/lexer"
	"github.com/v4nsh0x/pengu/parser"
)

const version = "0.1.2"

const logo = `
  🐧 Pengu v%s
  A fun, fast, and friendly programming language
`

const usage = `
Usage:
  pengu <file.pen>          Run a Pengu script
  pengu run <file.pen>      Run a Pengu script
  pengu repl                Start interactive REPL
  pengu build <file> -o <out>  Compile to executable
  pengu install <module>    Download and install a module
  pengu update              Update Pengu to the latest version
  pengu version             Show version
  pengu help                Show this help
`

// Run is the main entry point for the Pengu CLI.
func Run(args []string) {
	if len(args) < 2 {
		printHelp()
		return
	}

	command := args[1]

	switch command {
	case "version", "--version", "-v":
		fmt.Printf(logo, version)

	case "help", "--help", "-h":
		printHelp()

	case "repl":
		runREPL()

	case "run":
		if len(args) < 3 {
			fmt.Println("Error: missing file argument")
			fmt.Println("Usage: pengu run <file.pen>")
			os.Exit(1)
		}
		runFile(args[2])

	case "build":
		handleBuild(args[2:])

	case "install":
		if len(args) < 3 {
			fmt.Println("Error: missing module argument")
			fmt.Println("Usage: pengu install <module_name>")
			os.Exit(1)
		}
		handleInstall(args[2])

	case "update":
		handleUpdate()

	default:
		// If the argument ends with .pen, treat it as a file to run
		if strings.HasSuffix(command, ".pen") {
			// Check for -o flag (pengu file.pen -o output)
			if len(args) >= 4 && args[2] == "-o" {
				handleBuild([]string{command, "-o", args[3]})
			} else {
				runFile(command)
			}
		} else {
			fmt.Printf("Unknown command: %s\n", command)
			printHelp()
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Printf(logo, version)
	fmt.Println(usage)
}

func runFile(filename string) {
	interp := interpreter.New()
	err := interp.RunFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runREPL() {
	fmt.Printf(logo, version)
	fmt.Println("  Type 'exit' or press Ctrl+C to quit\n")

	interp := interpreter.New()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("🐧 > ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println("Bye! 🐧")
			break
		}

		err := interp.Run(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func handleBuild(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: missing file argument")
		fmt.Println("Usage: pengu build <file.pen> -o <output>")
		os.Exit(1)
	}

	inputFile := args[0]
	outputName := strings.TrimSuffix(filepath.Base(inputFile), ".pen")

	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputName = args[i+1]
		}
	}

	// Auto-add platform-appropriate extension if none provided
	if !strings.Contains(filepath.Base(outputName), ".") {
		outputName += execExtension()
	}

	// Read the source file
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read file '%s'\n", inputFile)
		os.Exit(1)
	}

	// Verify it parses correctly (parse-only, no execution)
	l := lexer.New(string(source))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Build failed — source has errors:")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	p := parser.New(tokens)
	_, err = p.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Build failed — source has errors:")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Generate a Go wrapper that embeds the source
	err = generateBinary(string(source), outputName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Build Error:\n%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Built: %s\n", outputName)
}

func generateBinary(source, outputName string) error {
	// Create a temp directory for the build
	tmpDir, err := os.MkdirTemp("", "pengu-build-*")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	// We embed the source into a simple Go program that uses the Pengu interpreter
	escapedSource := strings.ReplaceAll(source, "\\", "\\\\")
	escapedSource = strings.ReplaceAll(escapedSource, "`", "` + \"`\" + `")

	goSource := fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/v4nsh0x/pengu/interpreter"
)

func main() {
	source := `+"`"+`%s`+"`"+`
	interp := interpreter.New()
	err := interp.Run(source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, escapedSource)

	// Write the Go source
	mainFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(mainFile, []byte(goSource), 0644)
	if err != nil {
		return fmt.Errorf("could not write build file: %s", err)
	}

	// Detect Go version dynamically
	goVer := runtime.Version() // e.g. "go1.25.5"
	goVer = strings.TrimPrefix(goVer, "go")

	goMod := fmt.Sprintf(`module pengu-build

go %s
`, goVer)

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		return fmt.Errorf("could not write go.mod: %s", err)
	}

	// Run go get to fetch the exact pengu interpreter version
	penguPkg := fmt.Sprintf("github.com/v4nsh0x/pengu@v%s", version)
	getCmd := exec.Command("go", "get", penguPkg)
	getCmd.Dir = tmpDir
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch pengu runtime (%s): %s", penguPkg, err)
	}

	// Build the binary
	absOutput, _ := filepath.Abs(outputName)
	cmd := exec.Command("go", "build", "-o", absOutput, ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// execExtension returns the correct executable file extension for the target OS.
// It checks the GOOS env var first (for cross-compilation), then falls back to runtime.GOOS.
func execExtension() string {
	targetOS := os.Getenv("GOOS")
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if targetOS == "windows" {
		return ".exe"
	}
	return ""
}

// handleInstall downloads a module from the official repo and saves it to the local modules directory.
func handleInstall(module string) {
	if !strings.HasSuffix(module, ".pen") {
		module += ".pen"
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/v4nsh0x/pengu/main/modules/%s", module)

	fmt.Printf("Downloading %s...\n", module)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading module: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Error: Module '%s' not found (status %d)\n", module, resp.StatusCode)
		os.Exit(1)
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error resolving executable path: %v\n", err)
		os.Exit(1)
	}
	modulesDir := filepath.Join(filepath.Dir(execPath), "modules")

	err = os.MkdirAll(modulesDir, 0755)
	if err != nil {
		fmt.Printf("Error creating modules directory: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(modulesDir, module)
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully installed %s to %s\n", module, outPath)
}
