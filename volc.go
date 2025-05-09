package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

func main() {
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("API_KEY"),
		arkruntime.WithBaseUrl(os.Getenv("BASE_URL")),
	)

	fmt.Println("----- function call mulstiple rounds request -----")
	ctx := context.Background()
	// Step 1: send the conversation and available functions to the model
	req := model.CreateChatCompletionRequest{
		Model: os.Getenv("MODEL"),
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("你是世界上最棒的人工智能助手"),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("北京的天气怎么样？"),
				},
			},
		},
		Tools: []*model.Tool{
			{
				Type: model.ToolTypeFunction,
				Function: &model.FunctionDefinition{
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
								"description": "枚举值有celsius、fahrenheit",
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
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("chat error: %v\n", err)
		return
	}
	// extend conversation with assistant's reply
	req.Messages = append(req.Messages, &resp.Choices[0].Message)

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
			&model.ChatCompletionMessage{
				Role:       model.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
				Content: &model.ChatCompletionMessageContent{
					StringValue: &functionResponse,
				},
			},
		)
	}
	// get a new response from the model where it can see the function response
	secondResp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("second chat error: %v\n", err)
		return
	}
	fmt.Println("conversation", MustMarshal(req.Messages))
	fmt.Println("new message", MustMarshal(secondResp.Choices[0].Message))
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
		return `{"location": "Beijing", "temperature": "10", "unit": unit}`
	case "北京":
		return `{"location": "Beijing", "temperature": "10", "unit": unit}`
	case "shanghai":
		return `{"location": "Shanghai", "temperature": "23", "unit": unit})`
	case "上海":
		return `{"location": "Shanghai", "temperature": "23", "unit": unit})`
	default:
		return fmt.Sprintf(`{"location": %s, "temperature": "unknown"}`, location)
	}
}
func MustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
