# Shell Expansion in Case Files

Case JSON files can now contain shell command substitutions like `$(...)` which are evaluated at runtime via bash.

## Use Case

Embed dynamic content (e.g., base64-encoded files, API responses) directly in case JSON files without pre-processing.

## Syntax

Use standard bash command substitution in JSON string values:

```json
{
    "contents": [{
        "parts": [{
            "inlineData": {
                "mimeType": "application/pdf",
                "data": "$(curl -s https://example.com/doc.pdf | base64 -w0)"
            }
        }]
    }]
}
```

## Behavior

1. **Detection** - When loading a case file, check if raw content contains `$(`
2. **Preserve raw body** - Keep the original JSON string with `$(...)` intact
3. **Execute via bash** - Instead of direct curl exec, run through `bash -c` with heredoc:
   ```bash
   bash -c 'curl ... -d "$(cat <<FASTLLMCURL_EOF
   {json with $(cmd) here}
   FASTLLMCURL_EOF
   )"'
   ```
4. **Shell expansion** - Bash evaluates `$(...)` before passing to curl

## Implementation

| File | Changes |
|------|---------|
| `body.go` | Added `LoadCaseBodyRaw()`, `ApplyModelOverrideRaw()`, `ApplyStreamOverrideRaw()`, `ApplyJSONPatchRaw()` |
| `cli.go` | Added `BuildCurlCommandWithRawBody()`, `ExecCurlViaBash()`, `buildBashCommand()`, `shellQuote()` |
| `display.go` | Updated `StreamDisplay` to support bash execution for `--display` mode |
| `main.go` | Added `needsShellExpansion` detection and routing logic |

## Limitations

- `--patch` with shell expansion: JSON merge-patch will re-encode JSON, destroying `$(...)` in patched fields. Only top-level patches work correctly.
- Model/stream overrides only replace existing fields via regex; they don't add new fields to raw body.
- The heredoc delimiter `FASTLLMCURL_EOF` must not appear in the JSON content.

## Examples

```bash
# PDF analysis with Gemini
fastllmcurl -p novita -t gemini -c pdf -m pa/gemini-3-flash-preview

# Image from URL
# cases/image/chat.json:
# {"messages":[{"role":"user","content":[
#   {"type":"image_url","image_url":{"url":"data:image/png;base64,$(curl -s https://example.com/img.png | base64 -w0)"}},
#   {"type":"text","text":"describe this image"}
# ]}]}
fastllmcurl -p novita -t chat -c image -m gpt-4o

# Environment variables also work
# {"model":"$(echo $MODEL_NAME)","messages":[...]}
```

## Security Note

Shell expansion executes arbitrary commands. Only use trusted case files.
