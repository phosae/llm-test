package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if handleCompletionCommands(os.Args[1:]) {
		return
	}

	opts, err := ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	provider, ok := config.Providers[opts.Provider]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown provider %q\n", opts.Provider)
		os.Exit(1)
	}

	var token string
	if provider.NeedsAuth() {
		token, err = provider.GetToken(opts.Provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting token: %v\n", err)
			os.Exit(1)
		}
	}

	var body map[string]interface{}
	if opts.Case != "" {
		body, err = LoadCaseBody(opts.CasesDir, opts.Case, opts.Type)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading case: %v\n", err)
			os.Exit(1)
		}
	} else {
		body = make(map[string]interface{})
	}

	ApplyModelOverride(body, opts.Model)
	ApplyStreamOverride(body, opts.Stream)

	body, err = ApplyJSONPatch(body, opts.Patch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error applying patch: %v\n", err)
		os.Exit(1)
	}

	model := GetModelFromBody(body)
	if model == "" && opts.Type == "gemini" {
		fmt.Fprintf(os.Stderr, "Error: model is required for gemini type\n")
		os.Exit(1)
	}

	url := provider.BuildURL(opts.Type, opts.Stream, model)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if provider.NeedsAuth() {
		headers["Authorization"] = "Bearer " + token
	}
	if provider.FusionHeader {
		headers["X-Fusion-Beta"] = "with-provider-detail-2026-07-11"
	}

	curlArgs, err := BuildCurlCommand(url, headers, body, opts.CurlArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building curl command: %v\n", err)
		os.Exit(1)
	}

	if opts.DryRun {
		PrintCurlCommand(curlArgs)
		return
	}

	if opts.Display {
		display := NewStreamDisplay(opts.Type)
		if err := display.Run(curlArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := ExecCurl(curlArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing curl: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`fastllmcurl - LLM API CLI Tool

Usage:
  fastllmcurl -p <provider> -t <type> [-c <case>] [options] [curl-args...]

Required flags:
  -p <provider>    Provider name (novita, novita-dev, ppio, ppio-dev, or custom)
  -t <type>        Request type: chat, message, gemini, or response

Options:
  -c <case>        Case name (directory containing request body files)
  -m <model>       Override model in request body
  --stream         Enable streaming mode
  --display        Display streamed content as readable text (requires --stream)
  --patch <json>   JSON merge-patch to apply to request body
  --cases-dir      Directory containing cases (default: current directory)
  --dry-run        Print curl command without executing

All other arguments are passed through to curl.
When -d is provided via curl args, case body is ignored.

Examples:
  fastllmcurl -p novita -c hello -t chat
  fastllmcurl -p ppio -c hello -t message -m claude-sonnet-4-20250514
  fastllmcurl -p novita -c hello -t gemini -m gemini-pro --stream
  fastllmcurl -p novita -c hello -t response -m gpt-4o
  fastllmcurl -p novita -c hello -t chat --patch '{"temperature": 0.5}' -v
  fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[...]}'

Shell Completion:
  fastllmcurl completion bash  # output bash completion script
  fastllmcurl completion zsh   # output zsh completion script`)
}
