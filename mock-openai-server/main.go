package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	port       = flag.Int("port", 8888, "Port to listen on")
	fixedDelay = flag.Duration("delay", 0, "delay the response by this duration")
)

type ChatCompletionRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	Stream           bool      `json:"stream,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	N                int       `json:"n,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	User             string    `json:"user,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatCompletionChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

type ChunkChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func parseDelay(delayStr string) (time.Duration, error) {
	return time.ParseDuration(strings.ToLower(delayStr))
}

var ErrorModels = map[int]string{
	400: "Your credit balance is too low to access the API",
	403: "Your account has an outstanding balance. Please settle it to regain access.",
	429: "You exceeded your current quota, please check your plan and billing details.",
	500: "Internal server error",
	503: "The server is currently unavailable",
}

func MapError(code int, errMsg string) ErrorResponse {
	var codeStr = strconv.Itoa(code)
	if code >= 400 && code < 500 {
		codeStr = "rate_limit_exceeded"
	}
	if code >= 500 {
		codeStr = "server_error"
	}
	return ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}{
			Message: errMsg,
			Code:    codeStr,
		},
	}
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse delay from query parameter
	delayStr := r.URL.Query().Get("delay")
	var delay time.Duration
	var err error
	if delayStr != "" {
		delay, err = parseDelay(delayStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid delay format: %v", err), http.StatusBadRequest)
			return
		}
	} else if *fixedDelay > 0 {
		delay = *fixedDelay
	}

	// Parse request body
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Log request information
	log.Printf("Request received - RemoteAddr: %s, Method: %s, URL: %s, Model: %s", r.RemoteAddr, r.Method, r.URL.String(), req.Model)

	if code, err := strconv.Atoi(req.Model); err == nil {
		if errMsg, ok := ErrorModels[code]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(MapError(code, errMsg))
			return
		}
	}

	// Set default model if not provided
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}

	// Set default stream value
	if req.Stream {
		handleStreamingResponse(w, req, delay)
	} else {
		handleNonStreamingResponse(w, req, delay)
	}
}

func handleNonStreamingResponse(w http.ResponseWriter, req ChatCompletionRequest, delay time.Duration) {
	// Sleep for the specified delay
	if delay > 0 {
		log.Printf("Non-streaming request: sleeping for %v", delay)
		time.Sleep(delay)
	}

	// Create mock response
	response := ChatCompletionResponse{
		ID:      "chatcmpl-" + fmt.Sprintf("%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "This is a mock response from the OpenAI API server. The request was processed successfully.",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleStreamingResponse(w http.ResponseWriter, req ChatCompletionRequest, delay time.Duration) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create a flusher to ensure data is sent immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	completionID := "chatcmpl-" + fmt.Sprintf("%d", time.Now().Unix())
	created := time.Now().Unix()

	// Send initial chunks immediately
	chunks := []string{
		"This is a mock",
		" streaming response",
		" from the OpenAI",
		" API server. ",
	}

	for _, chunk := range chunks {
		chunkResponse := ChatCompletionChunk{
			ID:      completionID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []ChunkChoice{
				{
					Index: 0,
					Delta: Message{
						Role:    "assistant",
						Content: chunk,
					},
				},
			},
		}

		data, _ := json.Marshal(chunkResponse)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		// Small delay between chunks
		time.Sleep(100 * time.Millisecond)
	}

	// Sleep for the specified delay
	if delay > 0 {
		log.Printf("Streaming request: sleeping for %v before final chunk", delay)
		time.Sleep(delay)
	}

	// Send final chunk
	finalChunk := ChatCompletionChunk{
		ID:      completionID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []ChunkChoice{
			{
				Index: 0,
				Delta: Message{
					Role:    "assistant",
					Content: "The request was processed successfully with the specified delay.",
				},
				FinishReason: stringPtr("stop"),
			},
		},
	}

	data, _ := json.Marshal(finalChunk)
	fmt.Fprintf(w, "data: %s\n\n", data)

	// Send done signal
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func stringPtr(s string) *string {
	return &s
}

func main() {
	flag.Parse()

	// Create mux for routing
	mux := http.NewServeMux()

	// Register handlers for both endpoints
	mux.HandleFunc("/chat/completions", handleChatCompletions)
	mux.HandleFunc("/v1/chat/completions", handleChatCompletions)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Add root endpoint with usage information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"endpoints": map[string]string{
				"chat_completions":    "POST /chat/completions",
				"v1_chat_completions": "POST /v1/chat/completions",
				"health":              "GET /health",
			},
			"delay": "?delay=<duration>",
		})
	})

	addr := ":" + strconv.Itoa(*port)
	log.Printf("Starting mock OpenAI API server on %s, fixed delay: %v", addr, *fixedDelay)
	log.Printf("Available endpoints:")
	log.Printf("  POST /chat/completions")
	log.Printf("  POST /v1/chat/completions")
	log.Printf("  GET  /health")
	log.Printf("  GET  /")
	log.Printf("Use ?delay=5s to add delays")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
