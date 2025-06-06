# LLM Feature Test

Test OpenAI compatible API features
1. function calling
2. vision

## Usage

```bash
# Test all (default)
API_KEY=<key> BASE_URL=<url> MODEL=<model> go run .

# Test specific: -test f,v,f or -test v
go run . -test f    # function only
go run . -test v    # vision only
go run . -h         # help
```

## Environment Variables
- `API_KEY` - Your API key
- `BASE_URL` - Base URL  
- `MODEL` - Model name
