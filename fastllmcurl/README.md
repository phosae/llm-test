# fastllmcurl

A CLI wrapper around curl for LLM API calls.

## Install

```bash
go build -o fastllmcurl .
sudo cp fastllmcurl /usr/local/bin/
```

## Usage

```bash
fastllmcurl -p <provider> -t <type> [-c <case>] [curl-args...]
```

## Examples

```bash
# Using case file
fastllmcurl -p novita -c hello -t chat

# Direct body
fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hi"}]}'

# With model override
fastllmcurl -p ppio -c hello -t message -m claude-sonnet-4-20250514
```

## Setup

1. Create token file: `~/.llm-test/novita`
2. Create case files: `./hello/chat`, `./hello/message`, `./hello/gemini`

## Shell Completion

```bash
# Bash
fastllmcurl completion bash >> ~/.bashrc

# Zsh
fastllmcurl completion zsh >> ~/.zshrc
```

If you rename the binary (e.g., to `fcurl`), completion auto-detects the name:

```bash
# After renaming to fcurl
fcurl completion zsh >> ~/.zshrc
```

See [PROPOSAL.md](PROPOSAL.md) for full documentation.
