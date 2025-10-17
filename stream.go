package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

func stream(ctx context.Context, client *openai.Client) {
	fmt.Println("----- Stream Chat Completion Test -----")

	req := openai.ChatCompletionRequest{
		Model: os.Getenv("MODEL"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一位精通中国古典文学的学者，能够深入分析经典作品的哲学内涵。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "从《金瓶梅》《红楼梦》《水浒传》来看，人生的意义是什么？",
			},
		},
		Stream: true,
	}

	fmt.Println("Streaming request:")
	fmt.Println(string(MustMarshal(req)))
	fmt.Println("--------------------------------")

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("Stream creation error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Println("Streaming response:")
	fmt.Print("Assistant: ")

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println("\n\n[Stream completed]")
				break
			}
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			fmt.Print(content)
		}
	}

	fmt.Println("--------------------------------")
}
