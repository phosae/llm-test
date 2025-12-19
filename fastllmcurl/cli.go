package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Options struct {
	Provider string
	Case     string
	Type     string
	Model    string
	Stream   bool
	Display  bool
	Patch    string
	CasesDir string
	DryRun   bool
	CurlArgs []string
}

func ParseArgs(args []string) (*Options, error) {
	opts := &Options{
		CasesDir: "./cases",
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "-p":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-p requires a value")
			}
			opts.Provider = args[i+1]
			i += 2
		case "-c":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-c requires a value")
			}
			opts.Case = args[i+1]
			i += 2
		case "-t":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-t requires a value")
			}
			opts.Type = args[i+1]
			i += 2
		case "-m":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("-m requires a value")
			}
			opts.Model = args[i+1]
			i += 2
		case "--stream":
			opts.Stream = true
			i++
		case "--display":
			opts.Display = true
			i++
		case "--patch":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--patch requires a value")
			}
			opts.Patch = args[i+1]
			i += 2
		case "--cases-dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--cases-dir requires a value")
			}
			opts.CasesDir = args[i+1]
			i += 2
		case "--dry-run":
			opts.DryRun = true
			i++
		default:
			opts.CurlArgs = append(opts.CurlArgs, args[i:]...)
			i = len(args)
		}
	}

	if opts.Provider == "" {
		return nil, fmt.Errorf("-p (provider) is required")
	}
	if opts.Type == "" {
		return nil, fmt.Errorf("-t (type) is required")
	}
	if opts.Type != "chat" && opts.Type != "message" && opts.Type != "gemini" && opts.Type != "response" {
		return nil, fmt.Errorf("-t must be one of: chat, message, gemini, response")
	}
	if opts.Display && !opts.Stream {
		return nil, fmt.Errorf("--display requires --stream")
	}

	return opts, nil
}

func BuildCurlCommand(url string, headers map[string]string, body map[string]interface{}, extraArgs []string) ([]string, error) {
	return BuildCurlCommandWithRawBody(url, headers, body, "", extraArgs)
}

func BuildCurlCommandWithRawBody(url string, headers map[string]string, body map[string]interface{}, rawBody string, extraArgs []string) ([]string, error) {
	args := []string{"curl", "-X", "POST", url}

	for k, v := range headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
	}

	hasBody := false
	for _, arg := range extraArgs {
		if arg == "-d" || arg == "--data" || arg == "--data-raw" || arg == "--data-binary" {
			hasBody = true
			break
		}
	}

	if !hasBody {
		if rawBody != "" {
			args = append(args, "-d", rawBody)
		} else if len(body) > 0 {
			bodyJSON, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			args = append(args, "-d", string(bodyJSON))
		}
	}

	args = append(args, extraArgs...)

	return args, nil
}

func ExecCurl(args []string) error {
	return ExecCurlViaBash(args, false)
}

func ExecCurlViaBash(args []string, needsShellExpansion bool) error {
	if !needsShellExpansion {
		curlPath, err := exec.LookPath("curl")
		if err != nil {
			return fmt.Errorf("curl not found: %w", err)
		}
		return syscall.Exec(curlPath, args, os.Environ())
	}

	bashPath, err := exec.LookPath("bash")
	if err != nil {
		return fmt.Errorf("bash not found: %w", err)
	}

	cmdStr := buildBashCommand(args)
	return syscall.Exec(bashPath, []string{"bash", "-c", cmdStr}, os.Environ())
}

func buildBashCommand(args []string) string {
	var parts []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-d" && i+1 < len(args) {
			parts = append(parts, "-d", fmt.Sprintf("\"$(cat <<FASTLLMCURL_EOF\n%s\nFASTLLMCURL_EOF\n)\"", args[i+1]))
			i++
		} else {
			parts = append(parts, shellQuote(arg))
		}
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, c := range s {
		if c == ' ' || c == '"' || c == '\'' || c == '\\' || c == '$' || c == '`' || c == '!' || c == '{' || c == '}' || c == '[' || c == ']' || c == '(' || c == ')' || c == '<' || c == '>' || c == '|' || c == '&' || c == ';' || c == '*' || c == '?' || c == '~' || c == '#' {
			return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
		}
	}
	return s
}

func PrintCurlCommand(args []string) {
	var bodyJSON string
	var headers []string
	var otherArgs []string

	for i := 0; i < len(args); i++ {
		if args[i] == "-d" && i+1 < len(args) {
			bodyJSON = args[i+1]
			i++
		} else if args[i] == "-H" && i+1 < len(args) {
			headers = append(headers, args[i+1])
			i++
		} else {
			otherArgs = append(otherArgs, args[i])
		}
	}

	for i, arg := range otherArgs {
		if i > 0 {
			fmt.Print(" ")
		}
		if needsQuoting(arg) {
			fmt.Printf("'%s'", arg)
		} else {
			fmt.Print(arg)
		}
	}

	for _, h := range headers {
		fmt.Printf(" \\\n  -H '%s'", h)
	}

	if bodyJSON != "" {
		var prettyJSON []byte
		if strings.Contains(bodyJSON, "$(") {
			prettyJSON = []byte(bodyJSON)
		} else {
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(bodyJSON), &raw); err == nil {
				prettyJSON, _ = json.MarshalIndent(raw, "", "  ")
			} else {
				prettyJSON = []byte(bodyJSON)
			}
		}
		fmt.Printf(" \\\n  -d @- << EOF\n%s\nEOF\n", string(prettyJSON))
	} else {
		fmt.Println()
	}
}

func needsQuoting(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '"' || c == '\'' || c == '\\' || c == '{' || c == '}' || c == '[' || c == ']' || c == ':' || c == ',' {
			return true
		}
	}
	return false
}
