# LLM Feature Test

Test OpenAI compatible API features
1. function calling
2. vision
3. stream

## Usage

```bash
# Test all (default)
API_KEY=<key> BASE_URL=<url> MODEL=<model> go run .

# Test specific: -test f,v,s or individual
go run . -test f    # function only
go run . -test v    # vision only
go run . -test s    # stream only
go run . -h         # help
```

## Environment Variables
- `API_KEY` - Your API key
- `BASE_URL` - Base URL  
- `MODEL` - Model name
