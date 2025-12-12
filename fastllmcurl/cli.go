package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Options struct {
	Provider string
	Case     string
	Type     string
	Model    string
	Stream   bool
	Patch    string
	CasesDir string
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

	return opts, nil
}

func BuildCurlCommand(url string, headers map[string]string, body map[string]interface{}, extraArgs []string) ([]string, error) {
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

	if !hasBody && len(body) > 0 {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		args = append(args, "-d", string(bodyJSON))
	}

	args = append(args, extraArgs...)

	return args, nil
}

func ExecCurl(args []string) error {
	curlPath, err := exec.LookPath("curl")
	if err != nil {
		return fmt.Errorf("curl not found: %w", err)
	}

	return syscall.Exec(curlPath, args, os.Environ())
}
