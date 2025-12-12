# fastllmcurl

A CLI wrapper around curl for LLM API calls.

## Install

```bash
GOBIN=/usr/local/bin go install github.com/phosae/llm-test/fastllmcurl@latest
```

Or build from source:

```bash
go build -o fastllmcurl .
sudo cp fastllmcurl /usr/local/bin/
```

## Setup

```bash
git clone https://github.com/phosae/llm-test.git
mkdir -p ~/.llm-test
cp -r cases ~/.llm-test/
cp providers.yaml ~/.llm-test/
cp models.yaml ~/.llm-test/
echo "your-api-key" > ~/.llm-test/novita
```

then start play 🚀🚀🚀

```bash
fastllmcurl -p novita -t chat -m pa/gpt-5.2 -c hello
```

## Shell Completion

```bash
fastllmcurl completion bash >> ~/.bashrc
fastllmcurl completion zsh >> ~/.zshrc
```

See [PROPOSAL.md](PROPOSAL.md) for full documentation.

## Usage

```bash
fastllmcurl -p <provider> -t <type> [-c <case>] [options] [curl-args...]
```

## Options

| Flag | Description |
|------|-------------|
| `-p` | Provider (novita, ppio, local-fusion, etc.) |
| `-t` | Type: chat, message, gemini, response |
| `-c` | Case name |
| `-m` | Model override |
| `--stream` | Enable streaming |
| `--patch` | JSON merge-patch |
| `--dry-run` | Print curl command without executing |

## Cheat Sheet

Simple Examples

```bash
fastllmcurl -p novita -c hello -t chat
fastllmcurl -p ppio -c hello -t message -m pa/claude-sonnet-4-20250514
fastllmcurl -p novita -c hello -t chat --dry-run
fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[...]}'
```

Combine with cURL args and Unix Pipe to do Cache Testing

```shell
fastllmcurl -p ppio -t gemini -m pa/gmn-2.5-fls-lt -c cache -i | tee >(grep -i 'x-fusion-provider') | sed -n '/^{/,$p' | jq '.usageMetadata'
```

Patch Gemini protocol's think budget

```shell
fastllmcurl -p ppio -t gemini -m pa/gemini-3-pro-preview -c cache --patch '{"generationConfig": {"thinkingConfig": {"thinkingBudget": 128}} }'
```