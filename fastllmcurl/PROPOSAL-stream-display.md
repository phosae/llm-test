# Stream Display Mode Proposal

When `--stream` is enabled, add `--display` flag to parse SSE responses and display content in a chat-friendly format instead of raw JSON chunks.

## New Flag

| Flag | Description | Default |
|------|-------------|---------|
| `--display` | Display streamed content as readable text (requires `--stream`) | false |

## Behavior

When `--display` is enabled:

1. **Capture curl output** instead of using `syscall.Exec`, run curl as subprocess and read stdout
2. **Parse SSE events** - Read `data: {...}` lines from the stream
3. **Extract content deltas** based on API type:
   - `chat`: `.choices[0].delta.content`
   - `message`: `.delta.text` (from `content_block_delta` events)
   - `gemini`: `.candidates[0].content.parts[0].text`
   - `response`: `.delta` (output_text_delta events)
4. **Print incrementally** - Output text as it arrives, no newlines between chunks
5. **Handle completion** - Print newline when stream ends (`[DONE]` or `message_stop`)

## Content Extraction by Type

```
chat (OpenAI):
  event: {"choices":[{"delta":{"content":"Hello"}}]}
  extract: choices[0].delta.content

message (Claude):
  event: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}
  extract: delta.text (when type == "content_block_delta")

gemini (Google):
  event: {"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}
  extract: candidates[0].content.parts[0].text

response (OpenAI Responses API):
  event: {"type":"response.output_text.delta","delta":"Hello"}
  extract: delta (when type == "response.output_text.delta")
```

## Thinking/Reasoning Support

For models with thinking/reasoning (Claude, OpenAI o1/o3):

- `message`: Also extract `thinking` field from `content_block_delta` when `delta.type == "thinking_delta"`
- `chat`: Extract `reasoning_content` from `choices[0].delta.reasoning_content`
- `response`: Extract from `response.reasoning_summary_text.delta` events

Display format:
```
<thinking>
[thinking content here]
</thinking>

[regular content here]
```

## Examples

```bash
# Stream with display mode
fastllmcurl -p novita -c hello -t chat --stream --display

# Claude message with thinking
fastllmcurl -p novita -c thinking -t message --stream --display

# Gemini streaming
fastllmcurl -p novita -c hello -t gemini -m gemini-pro --stream --display
```

## Implementation Notes

- When `--display` is set, do NOT use `syscall.Exec`; instead use `exec.Command` with stdout pipe
- Buffer partial lines until newline received (SSE lines end with `\n\n`)
- Skip `data: [DONE]` lines
- Handle `event:` lines for type discrimination if needed
- Print to stdout unbuffered for real-time display
- On error events, print error message to stderr
