# LLM Feature Test

Test OpenAI compatible API features
1. function calling
2. vision
3. stream

## Usage

Environment Variables
- `API_KEY` - Your API key
- `BASE_URL` - Base URL  
- `MODEL` - Model name

```bash
# Test all (default)
API_KEY=<key> BASE_URL=<url> MODEL=<model> go run .

# Test specific: -test f,v,s or individual
go run . -test f    # function only
go run . -test v    # vision only
go run . -test s    # stream only
go run . -h         # help
```
### Docker image

Build image locally

```bash
#   install ko
#   GOBIN=/usr/local/bin/ go install github.com/google/ko@v0.18.0
    
KO_DOCKER_REPO=local ko build --local -B --platform linux/arm64 .
```

run in Docker

```bash
docker run -e API_KEY=<key> -e BASE_URL=<url> -e MODEL=<model> local/llm-test
```
