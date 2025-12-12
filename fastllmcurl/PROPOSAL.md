# fastllmcurl - LLM API CLI Tool

A Go-based CLI wrapper around curl that simplifies LLM API calls with built-in provider configurations and request templates.

## CLI Flags

| Flag | Description | Required |
|------|-------------|----------|
| `-p` | Provider name | Yes |
| `-t` | Type: `chat`, `message`, `gemini` | Yes |
| `-c` | Case name (optional if `-d` provided) | No |
| `-m` | Model (overrides body) | No |
| `--stream` | Stream mode (overrides body) | No |
| `--patch` | JSON merge-patch | No |
| `--cases-dir` | Directory containing cases | No |

All other flags are passed through to curl.
When `-d` is provided via curl args, case body is ignored.

## Provider Configuration

Location: `$HOME/.fastllmcurl/providers.yaml`

Built-in providers (novita, novita-dev, ppio, ppio-dev) are defaults but can be overridden by user config.

```yaml
novita:
  base_url: https://api.novita.ai
  path:
    chat: "openai/v1/chat/completions"
    message: "anthropic/v1/messages"
    gemini: "gemini/v1/models/{model}:generateContent"
    gemini_stream: "gemini/v1/models/{model}:streamGenerateContent"
  fusion_header: true

novita-dev:
  base_url: https://dev-api.novita.ai
  path:
    chat: "openai/v1/chat/completions"
    message: "anthropic/v1/messages"
    gemini: "gemini/v1/models/{model}:generateContent"
    gemini_stream: "gemini/v1/models/{model}:streamGenerateContent"
  fusion_header: true

ppio:
  base_url: https://api.ppio.com
  path:
    chat: "openai/v1/chat/completions"
    message: "anthropic/v1/messages"
    gemini: "gemini/v1/models/{model}:generateContent"
    gemini_stream: "gemini/v1/models/{model}:streamGenerateContent"
  fusion_header: true

ppio-dev:
  base_url: https://dev-api.ppinfra.com
  path:
    chat: "openai/v1/chat/completions"
    message: "anthropic/v1/messages"
    gemini: "gemini/v1/models/{model}:generateContent"
    gemini_stream: "gemini/v1/models/{model}:streamGenerateContent"
  fusion_header: true

fusion:
  base_url: http://127.0.0.1:8000/fusion/v1
  path:
    chat: "{model}/v1/chat/completions"
    message: "{model}/v1/messages"
    gemini: "{model}:generateContent"
    gemini_stream: "{model}:streamGenerateContent"
  fusion_header: false
  token_cmd: "cat /tmp/fusion-token"
```

## API Key / Token Loading

Priority order (first match wins):

1. **token_cmd** - Execute shell command, use stdout as token
   ```yaml
   my-provider:
     token_cmd: "vault kv get -field=api_key secret/my-provider"
   ```

2. **Plain file** - `$HOME/.llm-test/{provider}`
   ```
   $HOME/.llm-test/novita
   ```

3. **JWT file** - `$HOME/.llm-test/{provider}.jwt`
   ```
   $HOME/.llm-test/novita.jwt
   ```

4. **Environment variable** - `{PROVIDER}_API_KEY`
   ```
   NOVITA_API_KEY=sk-xxx
   ```

### token_cmd Examples

```yaml
# Password manager
my-provider:
  token_cmd: "pass show llm/my-provider"

# AWS secrets
aws-provider:
  token_cmd: "aws secretsmanager get-secret-value --secret-id llm-key | jq -r .SecretString"

# Dynamic JWT refresh
jwt-provider:
  token_cmd: "curl -s https://auth.example.com/token | jq -r .access_token"
```

## Case Directory Structure

Location: Current working directory first, then fallback to `$HOME/.llm-test/cases/`

```
# Local cases (checked first)
./hello/
./search/
./thinking/
./tool/

# Home cases (fallback)
$HOME/.llm-test/cases/
├── hello/
│   ├── chat      # OpenAI request body
│   ├── message   # Claude request body
│   └── gemini    # Gemini request body
├── search/
│   ├── chat
│   ├── message
│   └── gemini
├── thinking/
│   ├── chat
│   ├── message
│   └── gemini
└── tool/
    ├── chat
    ├── message
    └── gemini
```

Each file contains the JSON request body for that protocol.

## Headers

Automatically added:

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer {token}` |
| `Content-Type` | `application/json` |
| `X-Fusion-Beta` | `with-provider-detail-2026-07-11` (when `fusion_header: true`) |

## URL Construction

1. Load case body from `{cases_dir}/{case}/{type}`
2. Apply `-m model` → set `model` field in body
3. Apply `-s` → set `stream: true` in body
4. Apply `--patch` JSON merge-patch
5. Select path template:
   - If `-t gemini` and `-s`: use `gemini_stream`
   - Otherwise: use `{type}` (chat/message/gemini)
6. Replace `{model}` placeholder in path with model from body
7. Build final URL: `{base_url}/{path}`

## Examples

### Basic Usage

```bash
# Simple chat request
fastllmcurl -p novita -c hello -t chat

# Claude message
fastllmcurl -p ppio -c hello -t message

# Gemini
fastllmcurl -p novita -c hello -t gemini
```

### With Model Override

```bash
# Override model in request
fastllmcurl -p ppio -c hello -t chat -m gpt-4o

# Claude with specific model
fastllmcurl -p novita -c thinking -t message -m claude-sonnet-4-20250514
```

### Streaming

```bash
# Stream chat response
fastllmcurl -p novita -c hello -t chat -s

# Stream gemini (uses streamGenerateContent endpoint)
fastllmcurl -p ppio -c hello -t gemini -m gemini-pro -s
```

### JSON Merge-Patch

```bash
# Override temperature
fastllmcurl -p novita -c hello -t chat --patch '{"temperature": 0.5}'

# Override multiple fields
fastllmcurl -p ppio -c hello -t message --patch '{"model": "claude-3-opus", "max_tokens": 4096}'
```

### Curl Passthrough

```bash
# Verbose output
fastllmcurl -p novita -c hello -t chat -v

# With timeout
fastllmcurl -p ppio -c hello -t message --connect-timeout 30

# Multiple curl options
fastllmcurl -p novita -c hello -t chat -v --stream --max-time 60
```

### Direct Body via Curl Args

```bash
# Skip case file, provide body directly
fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'

# With model in path for gemini
fastllmcurl -p novita -t gemini -m gemini-pro -d '{"contents":[{"role":"user","parts":[{"text":"Hello"}]}]}'
```

### Fusion Provider

```bash
# Model in URL path
fastllmcurl -p fusion -c hello -t chat -m gpt-4o
# URL: http://127.0.0.1:8000/fusion/v1/gpt-4o/v1/chat/completions

fastllmcurl -p fusion -c hello -t gemini -m gemini-pro -s
# URL: http://127.0.0.1:8000/fusion/v1/gemini-pro:streamGenerateContent
```

## Implementation Flow

1. Parse CLI flags, separate fastllmcurl flags from curl flags
2. Load built-in provider defaults
3. Merge user config from `$HOME/.fastllmcurl/providers.yaml`
4. Get provider config by name
5. Load API token (token_cmd → file → jwt → env)
6. Load case body from `{cases_dir}/{case}/{type}` (skip if `-c` not provided)
7. Apply `-m` model override to body
8. Apply `-s` stream override to body
9. Apply `--patch` JSON merge-patch to body
10. Select path template (handle gemini_stream for streaming gemini)
11. Replace `{model}` in path with model from body
12. Build URL: `{base_url}/{path}`
13. Build headers array
14. Execute curl with: URL, headers, body, and passthrough args

## Implementation Language

Go - for easy distribution as single binary.

## File Locations Summary

| Purpose | Location |
|---------|----------|
| Provider config | `$HOME/.fastllmcurl/providers.yaml` |
| API keys | `$HOME/.llm-test/{provider}` |
| JWT tokens | `$HOME/.llm-test/{provider}.jwt` |
| Cases | `./{case}/{type}` or `$HOME/.llm-test/cases/{case}/{type}` |
