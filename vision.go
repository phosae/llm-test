package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

func vision(ctx context.Context, client *openai.Client) {
	fmt.Println("----- OpenAI Vision API Test -----")

	// Test 1: Base64 format
	fmt.Println("\n=== Test 1: Base64 Format ===")
	testVisionBase64(ctx, client)

	// Test 2: URL format (assuming you have the image accessible via URL)
	fmt.Println("\n=== Test 2: URL Format ===")
	testVisionURL(ctx, client)
}

func testVisionBase64(ctx context.Context, client *openai.Client) {
	// Read and encode the image to base64

	imgPath := "./kodata/lightning-bolts.jpg"
	if koDataDir := os.Getenv("KO_DATA_PATH"); koDataDir != "" {
		imgPath = fmt.Sprintf("%s/lightning-bolts.jpg", koDataDir)
	}
	imageData, err := readImageAsBase64(imgPath)
	if err != nil {
		fmt.Printf("Error reading image: %v\n", err)
		return
	}

	req := openai.ChatCompletionRequest{
		Model: os.Getenv("MODEL"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "What do you see in this image? Please describe the weather phenomenon, lighting conditions, and any notable features in detail.",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    fmt.Sprintf("data:image/jpeg;base64,%s", imageData),
							Detail: openai.ImageURLDetailHigh,
						},
					},
				},
			},
		},
		MaxTokens: 4096,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Vision API error (base64): %v\n", err)
		return
	}

	fmt.Println("Base64 Response:")
	fmt.Println(string(MustMarshal(resp.Choices)))
	if len(resp.Choices) > 0 {
		fmt.Printf("\nVision Analysis (Base64): %s\n", resp.Choices[0].Message.Content)
	}
}

func testVisionURL(ctx context.Context, client *openai.Client) {
	imageURL := "https://images.nationalgeographic.org/image/upload/t_edhub_resource_key_image/v1638886301/EducationHub/photos/lightning-bolts.jpg"

	req := openai.ChatCompletionRequest{
		Model: os.Getenv("MODEL"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Analyze this lightning photograph. What can you tell me about the atmospheric conditions, the type of lightning, and the overall scene?",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL:    imageURL,
							Detail: openai.ImageURLDetailHigh,
						},
					},
				},
			},
		},
		MaxTokens: 4096,
	}

	fmt.Println("URL Request:")
	fmt.Println(string(MustMarshal(req)))
	fmt.Println("--------------------------------")

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Vision API error (URL): %v\n", err)
		return
	}

	fmt.Println("URL Response:")
	fmt.Println(string(MustMarshal(resp.Choices)))
	if len(resp.Choices) > 0 {
		fmt.Printf("\nVision Analysis (URL): %s\n", resp.Choices[0].Message.Content)
	}
}

// readImageAsBase64 reads an image file and returns it as a base64 encoded string
func readImageAsBase64(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	base64String := base64.StdEncoding.EncodeToString(imageData)
	return base64String, nil
}
