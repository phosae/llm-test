# Mock OpenAI API Server

A simple mock server that simulates the OpenAI Chat Completions API with configurable delays for testing purposes.

## Features

- Serves both `/chat/completions` and `/v1/chat/completions` endpoints
- Supports both streaming and non-streaming responses
- Configurable delays via query parameter
- Health check endpoint
- Web interface with usage information

## Usage

### Starting the Server

```bash
# Start on default port 8080
go run main.go

# Start on custom port
go run main.go -port 9000
```

### API Endpoints

#### Chat Completions (Non-streaming)
```bash
curl -X POST http://localhost:8080/chat/completions?delay=5s \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }'
```

#### Chat Completions (Streaming)
```bash
curl -X POST http://localhost:8080/v1/chat/completions?delay=1m \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  }'
```

### Delay Options

- `5s` - 5 seconds delay
- `1m` - 1 minute delay  
- `1h` - 1 hour delay
- Any valid Go duration (e.g., `30s`, `2m30s`)

### Behavior

#### Non-streaming Mode
- Sleeps for the specified delay
- Returns a complete JSON response

#### Streaming Mode
- Sends initial chunks immediately
- Sleeps for the specified delay
- Sends final chunk with completion signal

### Health Check

```bash
curl http://localhost:8080/health
```

### Web Interface

Visit `http://localhost:8080/` for a web interface with usage information and examples.

## Testing

Run the test client to verify the server functionality:

```bash
# In a separate terminal
go run test_client.go
```

## Example Responses

### Non-streaming Response
```json
{
  "id": "chatcmpl-1234567890",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-3.5-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "This is a mock response from the OpenAI API server. The request was processed successfully."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### Streaming Response
Each chunk follows the format:
```
data: {"id":"chatcmpl-1234567890","object":"chat.completion.chunk","created":1234567890,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"role":"assistant","content":"chunk content"}}]}

data: [DONE]
``` 