package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PathConfig struct {
	Chat         string `yaml:"chat"`
	Message      string `yaml:"message"`
	Gemini       string `yaml:"gemini"`
	GeminiStream string `yaml:"gemini_stream"`
	Response     string `yaml:"response"`
}

type Provider struct {
	BaseURL      string     `yaml:"base_url"`
	Path         PathConfig `yaml:"path"`
	FusionHeader bool       `yaml:"fusion_header"`
	TokenCmd     string     `yaml:"token_cmd"`
	AuthHeader   *bool      `yaml:"auth_header"`
	Models       []string   `yaml:"models"`
	ModelsRef    []ModelRef `yaml:"models_ref"`
}

type ModelRef struct {
	Ref    string `yaml:"ref"`
	Prefix string `yaml:"prefix"`
}

type Config struct {
	Providers  map[string]*Provider
	ModelLists map[string][]string `yaml:"model_lists"`
}

func boolPtr(b bool) *bool {
	return &b
}

var builtinProviders = map[string]*Provider{
	"novita": {
		BaseURL: "https://api.novita.ai",
		Path: PathConfig{
			Chat:         "openai/v1/chat/completions",
			Message:      "anthropic/v1/messages",
			Gemini:       "gemini/v1/models/{model}:generateContent",
			GeminiStream: "gemini/v1/models/{model}:streamGenerateContent",
			Response:     "openai/v1/responses",
		},
		ModelsRef: []ModelRef{
			{
				Ref:    "openai",
				Prefix: "pa/",
			},
			{
				Ref:    "anthropic",
				Prefix: "pa/",
			},
			{
				Ref:    "google",
				Prefix: "pa/",
			},
		},
		FusionHeader: true,
	},
	"novita-dev": {
		BaseURL: "https://dev-api.novita.ai",
		Path: PathConfig{
			Chat:         "openai/v1/chat/completions",
			Message:      "anthropic/v1/messages",
			Gemini:       "gemini/v1/models/{model}:generateContent",
			GeminiStream: "gemini/v1/models/{model}:streamGenerateContent",
			Response:     "openai/v1/responses",
		},
		ModelsRef: []ModelRef{
			{
				Ref:    "openai",
				Prefix: "pa/",
			},
			{
				Ref:    "anthropic",
				Prefix: "pa/",
			},
			{
				Ref:    "google",
				Prefix: "pa/",
			},
		},
		FusionHeader: true,
	},
	"ppio": {
		BaseURL: "https://api.ppio.com",
		Path: PathConfig{
			Chat:         "openai/v1/chat/completions",
			Message:      "anthropic/v1/messages",
			Gemini:       "gemini/v1/models/{model}:generateContent",
			GeminiStream: "gemini/v1/models/{model}:streamGenerateContent",
			Response:     "openai/v1/responses",
		},
		ModelsRef: []ModelRef{
			{
				Ref:    "openai",
				Prefix: "pa/",
			},
			{
				Ref:    "anthropic",
				Prefix: "pa/",
			},
			{
				Ref:    "google",
				Prefix: "pa/",
			},
		},
		FusionHeader: true,
	},
	"ppio-dev": {
		BaseURL: "https://dev-api.ppinfra.com",
		Path: PathConfig{
			Chat:         "openai/v1/chat/completions",
			Message:      "anthropic/v1/messages",
			Gemini:       "gemini/v1/models/{model}:generateContent",
			GeminiStream: "gemini/v1/models/{model}:streamGenerateContent",
			Response:     "openai/v1/responses",
		},
		ModelsRef: []ModelRef{
			{
				Ref:    "openai",
				Prefix: "pa/",
			},
			{
				Ref:    "anthropic",
				Prefix: "pa/",
			},
			{
				Ref:    "google",
				Prefix: "pa/",
			},
		},
		FusionHeader: true,
	},
	"local-fusion": {
		BaseURL: "http://localhost:8000/fusion/v1",
		Path: PathConfig{
			Chat:         "{model}/v1/chat/completions",
			Message:      "{model}/v1/messages",
			Gemini:       "{model}:generateContent",
			GeminiStream: "{model}:streamGenerateContent",
			Response:     "{model}/v1/responses",
		},
		ModelsRef: []ModelRef{
			{
				Ref: "openai",
			},
			{
				Ref: "anthropic",
			},
			{
				Ref: "google",
			},
		},
		FusionHeader: false,
		AuthHeader:   boolPtr(false),
	},
}

func LoadConfig() (*Config, error) {
	config := &Config{
		Providers:  make(map[string]*Provider),
		ModelLists: make(map[string][]string),
	}

	for name, p := range builtinProviders {
		provider := *p
		config.Providers[name] = &provider
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, nil
	}

	providersPath := filepath.Join(homeDir, ".llm-test", "providers.yaml")
	if data, err := os.ReadFile(providersPath); err == nil {
		var userProviders map[string]*Provider
		if err := yaml.Unmarshal(data, &userProviders); err != nil {
			return nil, fmt.Errorf("failed to parse providers.yaml: %w", err)
		}
		for name, p := range userProviders {
			if existing, ok := config.Providers[name]; ok {
				mergeProvider(existing, p)
			} else {
				config.Providers[name] = p
			}
		}
	}

	modelsPath := filepath.Join(homeDir, ".llm-test", "models.yaml")
	if data, err := os.ReadFile(modelsPath); err == nil {
		var modelLists map[string][]string
		if err := yaml.Unmarshal(data, &modelLists); err != nil {
			return nil, fmt.Errorf("failed to parse models.yaml: %w", err)
		}
		for name, models := range modelLists {
			config.ModelLists[name] = models
		}
	}

	return config, nil
}

func mergeProvider(dst, src *Provider) {
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.Path.Chat != "" {
		dst.Path.Chat = src.Path.Chat
	}
	if src.Path.Message != "" {
		dst.Path.Message = src.Path.Message
	}
	if src.Path.Gemini != "" {
		dst.Path.Gemini = src.Path.Gemini
	}
	if src.Path.GeminiStream != "" {
		dst.Path.GeminiStream = src.Path.GeminiStream
	}
	if src.Path.Response != "" {
		dst.Path.Response = src.Path.Response
	}
	if src.TokenCmd != "" {
		dst.TokenCmd = src.TokenCmd
	}
	if src.AuthHeader != nil {
		dst.AuthHeader = src.AuthHeader
	}
	if len(src.Models) > 0 {
		dst.Models = src.Models
	}
	if len(src.ModelsRef) > 0 {
		dst.ModelsRef = src.ModelsRef
	}
	dst.FusionHeader = src.FusionHeader
}

func (c *Config) GetModels(providerName string) []string {
	return c.getModelsWithVisited(providerName, make(map[string]bool))
}

func (c *Config) getModelsWithVisited(providerName string, visited map[string]bool) []string {
	if visited[providerName] {
		return nil
	}
	visited[providerName] = true

	provider, ok := c.Providers[providerName]
	if !ok {
		return nil
	}

	var result []string

	if len(provider.Models) > 0 {
		result = append(result, provider.Models...)
	}

	if m, ok := c.ModelLists[providerName]; ok {
		result = append(result, m...)
	}

	for _, ref := range provider.ModelsRef {
		var models []string
		if _, ok := c.Providers[ref.Ref]; ok {
			models = c.getModelsWithVisited(ref.Ref, visited)
		} else if m, ok := c.ModelLists[ref.Ref]; ok {
			models = m
		}
		result = append(result, applyPrefix(models, ref.Prefix)...)
	}

	return result
}

func applyPrefix(models []string, prefix string) []string {
	if prefix == "" || len(models) == 0 {
		return models
	}
	result := make([]string, len(models))
	for i, m := range models {
		result[i] = prefix + m
	}
	return result
}

func (p *Provider) GetToken(providerName string) (string, error) {
	var tried []string

	if p.TokenCmd != "" {
		cmd := exec.Command("sh", "-c", p.TokenCmd)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("token_cmd %q failed: %w", p.TokenCmd, err)
		}
		return strings.TrimSpace(string(output)), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home dir: %w", err)
	}

	tokenFile := filepath.Join(homeDir, ".llm-test", providerName)
	if data, err := os.ReadFile(tokenFile); err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	tried = append(tried, tokenFile)

	jwtFile := filepath.Join(homeDir, ".llm-test", providerName+".jwt")
	if data, err := os.ReadFile(jwtFile); err == nil {
		return strings.TrimSpace(string(data)), nil
	}
	tried = append(tried, jwtFile)

	envKey := strings.ToUpper(strings.ReplaceAll(providerName, "-", "_")) + "_API_KEY"
	if token := os.Getenv(envKey); token != "" {
		return token, nil
	}
	tried = append(tried, "env:"+envKey)

	return "", fmt.Errorf("no token found for provider %q, tried: %v", providerName, tried)
}

func (p *Provider) GetPath(reqType string, stream bool) string {
	if reqType == "gemini" && stream {
		return p.Path.GeminiStream
	}
	switch reqType {
	case "chat":
		return p.Path.Chat
	case "message":
		return p.Path.Message
	case "gemini":
		return p.Path.Gemini
	case "response":
		return p.Path.Response
	default:
		return ""
	}
}

func (p *Provider) BuildURL(reqType string, stream bool, model string) string {
	path := p.GetPath(reqType, stream)
	path = strings.ReplaceAll(path, "{model}", model)
	return strings.TrimSuffix(p.BaseURL, "/") + "/" + path
}

func (p *Provider) NeedsAuth() bool {
	if p.AuthHeader == nil {
		return true
	}
	return *p.AuthHeader
}
