package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateBashCompletion(binName string) string {
	return fmt.Sprintf(`_%s() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "${prev}" in
        -p)
            local providers="novita novita-dev ppio ppio-dev local-fusion"
            local config_file="$HOME/.llm-test/providers.yaml"
            if [[ -f "$config_file" ]]; then
                local custom_providers=$(grep -E '^[a-zA-Z0-9_-]+:' "$config_file" 2>/dev/null | sed 's/:.*//')
                providers="$providers $custom_providers"
            fi
            COMPREPLY=( $(compgen -W "$providers" -- "$cur") )
            return 0
            ;;
        -c)
            local cases_dir="./cases"
            for ((i=1; i<COMP_CWORD; i++)); do
                if [[ "${COMP_WORDS[i]}" == "--cases-dir" ]]; then
                    cases_dir="${COMP_WORDS[i+1]}"
                    break
                fi
            done
            local cases=""
            if [[ -d "$cases_dir" ]]; then
                cases=$(find "$cases_dir" -maxdepth 1 -type d ! -name "." ! -name ".." -exec basename {} \; 2>/dev/null)
            fi
            if [[ -d "$HOME/.llm-test/cases" ]]; then
                cases="$cases $(find "$HOME/.llm-test/cases" -maxdepth 1 -type d ! -name "." ! -name ".." -exec basename {} \; 2>/dev/null)"
            fi
            COMPREPLY=( $(compgen -W "$cases" -- "$cur") )
            return 0
            ;;
        -t)
            COMPREPLY=( $(compgen -W "chat message gemini" -- "$cur") )
            return 0
            ;;
        -m)
            local provider=""
            for ((i=1; i<COMP_CWORD; i++)); do
                if [[ "${COMP_WORDS[i]}" == "-p" ]]; then
                    provider="${COMP_WORDS[i+1]}"
                    break
                fi
            done
            if [[ -n "$provider" ]]; then
                local models=$(%s __complete-models "$provider" 2>/dev/null)
                COMPREPLY=( $(compgen -W "$models" -- "$cur") )
            fi
            return 0
            ;;
        --patch|--cases-dir)
            return 0
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        opts="-p -c -t -m --stream --patch --cases-dir"
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
        return 0
    fi
}

complete -F _%s %s
`, binName, binName, binName, binName)
}

func generateZshCompletion(binName string) string {
	return fmt.Sprintf(`#compdef %s

_%s() {
    local curcontext="$curcontext" state line
    typeset -A opt_args
    
    _arguments -C \
        '-p[Provider name]:provider:->providers' \
        '-c[Case name]:case:->cases' \
        '-t[Request type]:type:->types' \
        '-m[Model override]:model:->models' \
        '--stream[Enable streaming]' \
        '--patch[JSON merge-patch]:patch:' \
        '--cases-dir[Cases directory]:directory:_files -/' \
        '*::curl args:' && return

    case "$state" in
        providers)
            local -a providers
            providers=(novita novita-dev ppio ppio-dev local-fusion)
            if [[ -f "$HOME/.llm-test/providers.yaml" ]]; then
                local custom_providers
                custom_providers=(${(f)"$(grep -E '^[a-zA-Z0-9_-]+:' "$HOME/.llm-test/providers.yaml" 2>/dev/null | sed 's/:.*//')"})
                providers+=($custom_providers)
            fi
            _describe -t providers 'provider' providers
            ;;
        cases)
            local -a cases
            local cases_dir="./cases"
            if [[ -d "$cases_dir" ]]; then
                cases=(${(f)"$(find "$cases_dir" -maxdepth 1 -type d ! -name "." ! -name ".." ! -name ".*" -exec basename {} \; 2>/dev/null)"})
            fi
            if [[ -d "$HOME/.llm-test/cases" ]]; then
                cases+=(${(f)"$(find "$HOME/.llm-test/cases" -maxdepth 1 -type d ! -name "." ! -name ".." ! -name ".*" -exec basename {} \; 2>/dev/null)"})
            fi
            _describe -t cases 'case' cases
            ;;
        types)
            local -a types
            types=(chat message gemini)
            _describe -t types 'type' types
            ;;
        models)
            local -a models
            local provider="${opt_args[-p]}"
            if [[ -n "$provider" ]]; then
                models=(${(f)"$(%s __complete-models "$provider" 2>/dev/null)"})
                _describe -t models 'model' models
            fi
            ;;
    esac
}

compdef _%s %s
`, binName, binName, binName, binName, binName)
}

func getBinaryName() string {
	exe, err := os.Executable()
	if err != nil {
		return "fcurl"
	}
	return filepath.Base(exe)
}

func printCompletionScript(shell string) {
	binName := getBinaryName()
	switch shell {
	case "bash":
		fmt.Print(generateBashCompletion(binName))
	case "zsh":
		fmt.Print(generateZshCompletion(binName))
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s. Supported: bash, zsh\n", shell)
		os.Exit(1)
	}
}

func listProviders(config *Config) {
	for name := range config.Providers {
		fmt.Println(name)
	}
}

func listCases(casesDir string) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			fmt.Println(entry.Name())
		}
	}
}

func listTypes() {
	fmt.Println("chat")
	fmt.Println("message")
	fmt.Println("gemini")
}

func handleCompletionCommands(args []string) bool {
	if len(args) < 1 {
		return false
	}

	switch args[0] {
	case "completion":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: fastllmcurl completion <bash|zsh>")
			os.Exit(1)
		}
		printCompletionScript(args[1])
		return true
	case "__complete-providers":
		config, _ := LoadConfig()
		if config != nil {
			listProviders(config)
		}
		return true
	case "__complete-cases":
		casesDir := "."
		if len(args) > 1 {
			casesDir = args[1]
		}
		listCases(casesDir)
		return true
	case "__complete-types":
		listTypes()
		return true
	case "__complete-models":
		if len(args) < 2 {
			return true
		}
		config, _ := LoadConfig()
		if config != nil {
			models := config.GetModels(args[1])
			for _, m := range models {
				fmt.Println(m)
			}
		}
		return true
	case "models":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: fastllmcurl models <provider>")
			os.Exit(1)
		}
		config, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		models := config.GetModels(args[1])
		for _, m := range models {
			fmt.Println(m)
		}
		return true
	}

	return false
}

func installCompletionHint() string {
	home, _ := os.UserHomeDir()
	return fmt.Sprintf(`To enable shell completion:

Bash:
  fastllmcurl completion bash > %s
  echo 'source %s' >> ~/.bashrc

Zsh:
  fastllmcurl completion zsh > %s
  echo 'source %s' >> ~/.zshrc
`,
		filepath.Join(home, ".fastllmcurl", "completion.bash"),
		filepath.Join(home, ".fastllmcurl", "completion.bash"),
		filepath.Join(home, ".fastllmcurl", "completion.zsh"),
		filepath.Join(home, ".fastllmcurl", "completion.zsh"),
	)
}
