package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type StreamDisplay struct {
	reqType             string
	inThinking          bool
	needsShellExpansion bool
}

func NewStreamDisplay(reqType string, needsShellExpansion bool) *StreamDisplay {
	return &StreamDisplay{reqType: reqType, needsShellExpansion: needsShellExpansion}
}

func (s *StreamDisplay) Run(args []string) error {
	var cmd *exec.Cmd
	if s.needsShellExpansion {
		cmdStr := buildBashCommand(args)
		cmd = exec.Command("bash", "-c", cmdStr)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start curl: %w", err)
	}

	s.processStream(stdout)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("curl failed: %w", err)
	}

	return nil
}

func (s *StreamDisplay) processStream(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		s.processLine(line)
	}
	if s.inThinking {
		fmt.Println("\n</thinking>")
	}
	fmt.Println()
}

func (s *StreamDisplay) processLine(line string) {
	if !strings.HasPrefix(line, "data: ") {
		return
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "[DONE]" {
		return
	}

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return
	}

	switch s.reqType {
	case "chat":
		s.extractChat(event)
	case "message":
		s.extractMessage(event)
	case "gemini":
		s.extractGemini(event)
	case "response":
		s.extractResponse(event)
	}
}

func (s *StreamDisplay) extractChat(event map[string]interface{}) {
	choices, ok := event["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return
	}
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return
	}

	if reasoning, ok := delta["reasoning_content"].(string); ok && reasoning != "" {
		if !s.inThinking {
			fmt.Print("<thinking>\n")
			s.inThinking = true
		}
		fmt.Print(reasoning)
	}

	if content, ok := delta["content"].(string); ok && content != "" {
		if s.inThinking {
			fmt.Print("\n</thinking>\n\n")
			s.inThinking = false
		}
		fmt.Print(content)
	}
}

func (s *StreamDisplay) extractMessage(event map[string]interface{}) {
	eventType, _ := event["type"].(string)

	if eventType == "content_block_delta" {
		delta, ok := event["delta"].(map[string]interface{})
		if !ok {
			return
		}

		deltaType, _ := delta["type"].(string)

		if deltaType == "thinking_delta" {
			if thinking, ok := delta["thinking"].(string); ok {
				if !s.inThinking {
					fmt.Print("<thinking>\n")
					s.inThinking = true
				}
				fmt.Print(thinking)
			}
		}

		if deltaType == "text_delta" {
			if text, ok := delta["text"].(string); ok {
				if s.inThinking {
					fmt.Print("\n</thinking>\n\n")
					s.inThinking = false
				}
				fmt.Print(text)
			}
		}
	}
}

func (s *StreamDisplay) extractGemini(event map[string]interface{}) {
	candidates, ok := event["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return
	}
	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return
	}
	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return
	}
	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return
	}
	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return
	}

	if thought, ok := part["thought"].(bool); ok && thought {
		if text, ok := part["text"].(string); ok {
			if !s.inThinking {
				fmt.Print("<thinking>\n")
				s.inThinking = true
			}
			fmt.Print(text)
		}
		return
	}

	if text, ok := part["text"].(string); ok {
		if s.inThinking {
			fmt.Print("\n</thinking>\n\n")
			s.inThinking = false
		}
		fmt.Print(text)
	}
}

func (s *StreamDisplay) extractResponse(event map[string]interface{}) {
	eventType, _ := event["type"].(string)

	if eventType == "response.reasoning_summary_text.delta" {
		if delta, ok := event["delta"].(string); ok {
			if !s.inThinking {
				fmt.Print("<thinking>\n")
				s.inThinking = true
			}
			fmt.Print(delta)
		}
	}

	if eventType == "response.output_text.delta" {
		if delta, ok := event["delta"].(string); ok {
			if s.inThinking {
				fmt.Print("\n</thinking>\n\n")
				s.inThinking = false
			}
			fmt.Print(delta)
		}
	}
}
