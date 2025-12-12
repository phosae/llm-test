# fastllmcurl - LLM API CLI Tool

A Go-based CLI wrapper around curl that simplifies LLM API calls with built-in provider configurations and request templates.

## CLI Flags

| Flag | Description | Required |
|------|-------------|----------|
| `-p` | Provider name | Yes |
| `-t` | Type: `chat`, `message`, `gemini`, `response` | Yes |
| `-c` | Case name (optional if `-d` provided) | No |
| `-m` | Model (overrides body) | No |
| `--stream` | Stream mode (overrides body) | No |
| `--patch` | JSON merge-patch | No |
| `--cases-dir` | Directory containing cases | No |
| `--dry-run` | Print curl command without executing | No |

All other flags are passed through to curl.
When `-d` is provided via curl args, case body is ignored.

## Provider Configuration

Location: `$HOME/.llm-test/providers.yaml`

Built-in providers (novita, novita-dev, ppio, ppio-dev, local-fusion) are defaults but can be overridden by user config.

```yaml
novita:
  base_url: https://api.novita.ai
  path:
    chat: "openai/v1/chat/completions"
    message: "anthropic/v1/messages"
    gemini: "gemini/v1/models/{model}:generateContent"
    gemini_stream: "gemini/v1/models/{model}:streamGenerateContent"
    response: "openai/v1/responses"
  fusion_header: true

local-fusion:
  base_url: http://localhost:8000/fusion/v1
  path:
    chat: "{model}/v1/chat/completions"
    message: "{model}/v1/messages"
    gemini: "{model}:generateContent"
    gemini_stream: "{model}:streamGenerateContent"
    response: "{model}/v1/responses"
  fusion_header: false
  auth_header: false
```

## Model Lists

Location: `$HOME/.llm-test/models.yaml`

```yaml
openai:
  - gpt-4o
  - gpt-4o-mini

anthropic:
  - claude-sonnet-4-20250514
  - claude-haiku-4-5-20251001

google:
  - gemini-2.5-flash-preview-05-20
  - gemini-2.5-pro-preview-06-05
```

Providers can reference model lists with optional prefix:

```yaml
ppio:
  models_ref:
    - ref: openai
      prefix: "pa/"
    - ref: anthropic
      prefix: "pa/"
```

## API Key / Token Loading

Priority order (first match wins):

1. **token_cmd** - Execute shell command, use stdout as token
2. **Plain file** - `$HOME/.llm-test/{provider}`
3. **JWT file** - `$HOME/.llm-test/{provider}.jwt`
4. **Environment variable** - `{PROVIDER}_API_KEY`

## Case Directory Structure

Location: Current working directory first, then fallback to `$HOME/.llm-test/cases/`

```
$HOME/.llm-test/cases/
├── hello/
│   ├── chat.json      # OpenAI request body
│   ├── message.json   # Claude request body
│   ├── gemini.json    # Gemini request body
│   └── response.json  # OpenAI Response API body
└── thinking/
    ├── chat.json
    └── message.json
```

## Headers

Automatically added:

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer {token}` (unless `auth_header: false`) |
| `Content-Type` | `application/json` |
| `X-Fusion-Beta` | `with-provider-detail-2026-07-11` (when `fusion_header: true`) |

## URL Construction

1. Load case body from `{cases_dir}/{case}/{type}.json`
2. Apply `-m model` → set `model` field in body
3. Apply `--stream` → set `stream: true` in body
4. Apply `--patch` JSON merge-patch
5. Select path template:
   - If `-t gemini` and `--stream`: use `gemini_stream`
   - Otherwise: use `{type}` (chat/message/gemini/response)
6. Replace `{model}` placeholder in path with model from body
7. Build final URL: `{base_url}/{path}`

## Examples

### Basic Usage

```bash
# Simple chat request
fastllmcurl -p novita -c hello -t chat

# Claude message
fastllmcurl -p ppio -c hello -t message

# OpenAI Response API
fastllmcurl -p novita -c hello -t response -m pa/gt-4p
```

### Dry Run

```bash
# Print curl command without executing
fastllmcurl -p novita -c hello -t chat --dry-run
```

### With Model Override

```bash
fastllmcurl -p ppio -c hello -t chat -m pa/gt-4p
fastllmcurl -p novita -c thinking -t message -m pa/claude-sonnet-4-20250514
```

### Streaming

```bash
fastllmcurl -p novita -c hello -t chat --stream
fastllmcurl -p ppio -c hello -t gemini -m gemini-pro --stream
```

### JSON Merge-Patch

```bash
fastllmcurl -p novita -c hello -t chat --patch '{"temperature": 0.5}'
```

### Curl Passthrough

```bash
fastllmcurl -p novita -c hello -t chat -v
fastllmcurl -p ppio -c hello -t message --connect-timeout 30
```

### Direct Body via Curl Args

```bash
fastllmcurl -p novita -t chat -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hello"}]}'
```

## File Locations Summary

| Purpose | Location |
|---------|----------|
| Provider config | `$HOME/.llm-test/providers.yaml` |
| Model lists | `$HOME/.llm-test/models.yaml` |
| API keys | `$HOME/.llm-test/{provider}` |
| JWT tokens | `$HOME/.llm-test/{provider}.jwt` |
| Cases | `./{case}/{type}.json` or `$HOME/.llm-test/cases/{case}/{type}.json` |
