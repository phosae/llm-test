package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func function(ctx context.Context, client *openai.Client) {
	fmt.Println("----- function call multiple rounds request -----")
	// Step 1: send the conversation and available functions to the model
	req := openai.ChatCompletionRequest{
		Model: os.Getenv("MODEL"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You're the best assistant in the world",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "What's the weather like in Beijing today?",
			},
		},
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. Beijing",
							},
							"unit": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"celsius", "fahrenheit"},
								"description": "Units the temperature will be returned in.",
							},
						},
						"required": []string{
							"location",
						},
					},
				},
			},
		},
	}

	fmt.Println("--------------------------------")
	fmt.Println("Round 1 request", string(MustMarshal(req)))
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("chat error: %v\n", err)
		return
	}
	fmt.Println("Round 1 response choices", string(MustMarshal(resp.Choices)))
	fmt.Println("--------------------------------")

	// extend conversation with assistant's reply
	req.Messages = append(req.Messages, resp.Choices[0].Message)

	// Step 2: check if the model wanted to call a function.
	// The model can choose to call one or more functions; if so,
	// the content will be a stringified JSON object adhering to
	// your custom schema (note: the model may hallucinate parameters).
	for _, toolCall := range resp.Choices[0].Message.ToolCalls {
		fmt.Println("calling function")
		fmt.Println("    id:", toolCall.ID)
		fmt.Println("    name:", toolCall.Function.Name)
		fmt.Println("    argument:", toolCall.Function.Arguments)
		functionResponse, err := CallAvailableFunctions(toolCall.Function.Name, toolCall.Function.Arguments)
		if err != nil {
			functionResponse = err.Error()
		}
		// extend conversation with function response
		req.Messages = append(req.Messages,
			openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    functionResponse,
				ToolCallID: toolCall.ID,
			},
		)
	}

	fmt.Println("--------------------------------")
	fmt.Println("Round 2 ReqMessages", MustMarshal(req.Messages))
	// get a new response from the model where it can see the function response
	secondResp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("second chat error: %v, resp: %v\n", err, secondResp)
		return
	}
	fmt.Println("Round 2 RespChoice", MustMarshal(secondResp.Choices))
}

func CallAvailableFunctions(name, arguments string) (string, error) {
	if name == "get_current_weather" {
		params := struct {
			Location string `json:"location"`
			Unit     string `json:"unit"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", fmt.Errorf("failed to parse function call name=%s arguments=%s", name, arguments)
		}
		return GetCurrentWeather(params.Location, params.Unit), nil
	} else {
		return "", fmt.Errorf("got unavailable function name=%s arguments=%s", name, arguments)
	}
}

// GetCurrentWeather get the current weather in a given location.
// Example dummy function hard coded to return the same weather.
// In production, this could be your backend API or an external API
func GetCurrentWeather(location, unit string) string {
	if unit == "" {
		unit = "celsius"
	}
	switch strings.ToLower(location) {
	case "beijing":
		return `{"location": "Beijing", "temperature": "5", "unit": "` + unit + `"}`
	case "北京":
		return `{"location": "Beijing", "temperature": "5", "unit": "` + unit + `"}`
	case "shanghai":
		return `{"location": "Shanghai", "temperature": "13", "unit": "` + unit + `"}`
	case "上海":
		return `{"location": "Shanghai", "temperature": "13", "unit": "` + unit + `"}`
	default:
		return fmt.Sprintf(`{"location": "%s", "temperature": "unknown"}`, location)
	}
}

func MustMarshal(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
