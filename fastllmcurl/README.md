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

## Examples

```bash
fastllmcurl -p novita -c hello -t chat
fastllmcurl -p ppio -c hello -t message -m claude-sonnet-4-20250514
fastllmcurl -p novita -c hello -t chat --dry-run
fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[...]}'
```
