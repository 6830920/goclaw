// Package ai provides AI model interfaces for Goclaw
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatCompletionRequest represents a request to a chat completion API
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatCompletionResponse represents a response from a chat completion API
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a choice in the response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client interface for AI model providers
type Client interface {
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

// ZhipuClient implements Client for Zhipu AI
type ZhipuClient struct {
	ApiKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewZhipuClient creates a new Zhipu AI client
func NewZhipuClient(apiKey, baseURL, model string) *ZhipuClient {
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	}
	
	if model == "" {
		model = "glm-4"
	}

	return &ZhipuClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to Zhipu AI
func (z *ZhipuClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = z.Model
	}

	// Prepare the request body
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", z.BaseURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+z.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := z.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Decode response
	var apiResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}

// SendMessage sends a simple message and gets a response
func (z *ZhipuClient) SendMessage(ctx context.Context, role, content string) (string, error) {
	req := ChatCompletionRequest{
		Model: z.Model,
		Messages: []Message{
			{Role: role, Content: content},
		},
		Stream: false,
	}

	resp, err := z.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// AnthropicMessage represents a chat message for Anthropic API
type AnthropicMessage struct {
	Role    string `json:"role"`    // "user", "assistant"
	Content string `json:"content"`
}

// AnthropicContent represents content in Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`  // Usually "text"
	Text string `json:"text"`
}

// AnthropicUsage represents token usage in Anthropic API
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicMessageRequest represents a request to an Anthropic-compatible API
type AnthropicMessageRequest struct {
	Model     string           `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int              `json:"max_tokens"`
	Stream    bool             `json:"stream"`
}

// AnthropicMessageResponse represents a response from an Anthropic-compatible API
type AnthropicMessageResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Model   string             `json:"model"`
	Content []AnthropicContent `json:"content"`
	Usage   AnthropicUsage     `json:"usage"`
}

// AnthropicCompatibleClient implements Client for Anthropic-compatible APIs like Minimax
type AnthropicCompatibleClient struct {
	ApiKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewAnthropicCompatibleClient creates a new client for Anthropic-compatible APIs
func NewAnthropicCompatibleClient(apiKey, baseURL, model string) *AnthropicCompatibleClient {
	if model == "" {
		model = "claude-3-sonnet-20240229" // default model
	}

	return &AnthropicCompatibleClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to an Anthropic-compatible API
func (a *AnthropicCompatibleClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Convert OpenAI format to Anthropic format
	anthropicReq := AnthropicMessageRequest{
		Model: a.Model,
		// Anthropic requires messages to alternate between user and assistant
		Messages:  convertToAnthropicMessages(req.Messages),
		MaxTokens: 4096, // Default max tokens for Anthropic
		Stream:    req.Stream,
	}

	// Prepare the request body
	requestBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// For Minimax, BaseURL should be like "https://api.minimaxi.com/anthropic", so we add "/v1/messages"
	endpoint := strings.TrimRight(a.BaseURL, "/") + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for Anthropic API
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.ApiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Make the request
	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Decode Anthropic response
	var anthropicResp AnthropicMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Anthropic response to OpenAI format
	openAIResp := &ChatCompletionResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   anthropicResp.Model,
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}

	// Convert content to choices
	if len(anthropicResp.Content) > 0 {
		var content string
		for _, c := range anthropicResp.Content {
			content += c.Text
		}

		choice := Choice{
			Index: 0,
			Message: Message{
				Role:    "assistant", // Anthropic returns "assistant" role
				Content: content,
			},
			FinishReason: "stop",
		}
		openAIResp.Choices = append(openAIResp.Choices, choice)
	}

	return openAIResp, nil
}

// convertToAnthropicMessages converts OpenAI messages to Anthropic format
func convertToAnthropicMessages(messages []Message) []AnthropicMessage {
	var anthropicMessages []AnthropicMessage
	
	for _, msg := range messages {
		// Anthropic requires role to be either "user" or "assistant"
		role := msg.Role
		if role == "system" {
			// Anthropic doesn't have a system role, so prepend to first user message
			// For simplicity, we'll treat system messages as user messages
			role = "user"
		}
		anthropicMessages = append(anthropicMessages, AnthropicMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	
	return anthropicMessages
}

// OpenAICompatibleClient implements Client for OpenAI-compatible APIs like Qwen
type OpenAICompatibleClient struct {
	ApiKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewOpenAICompatibleClient creates a new client for OpenAI-compatible APIs
func NewOpenAICompatibleClient(apiKey, baseURL, model string) *OpenAICompatibleClient {
	if model == "" {
		model = "gpt-3.5-turbo" // default model
	}

	return &OpenAICompatibleClient{
		ApiKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to an OpenAI-compatible API
func (o *OpenAICompatibleClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = o.Model
	}

	// Prepare the request body
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// For Qwen-Portal, BaseURL should be like "https://portal.qwen.ai/v1", so we add "/chat/completions"
	endpoint := strings.TrimRight(o.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+o.ApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := o.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Decode response
	var apiResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}

// SendMessage sends a simple message and gets a response
func (o *OpenAICompatibleClient) SendMessage(ctx context.Context, role, content string) (string, error) {
	req := ChatCompletionRequest{
		Model: o.Model,
		Messages: []Message{
			{Role: role, Content: content},
		},
		Stream: false,
	}

	resp, err := o.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	return resp.Choices[0].Message.Content, nil
}

// MultiProviderClient manages multiple AI providers and selects the appropriate one
type MultiProviderClient struct {
	Providers map[string]Client
}

// NewMultiProviderClient creates a new client that can handle multiple providers
func NewMultiProviderClient() *MultiProviderClient {
	return &MultiProviderClient{
		Providers: make(map[string]Client),
	}
}

// AddProvider adds a provider to the multi-provider client
func (m *MultiProviderClient) AddProvider(name string, client Client) {
	m.Providers[name] = client
}

// ChatCompletion makes a request using the appropriate provider
func (m *MultiProviderClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Determine which provider to use based on the model name
	providerName := ""
	if strings.Contains(strings.ToLower(req.Model), "minimax") || strings.Contains(strings.ToLower(req.Model), "minimax-m2") {
		providerName = "minimax"
	} else if strings.Contains(strings.ToLower(req.Model), "qwen") || strings.Contains(strings.ToLower(req.Model), "coder-model") {
		providerName = "qwen"
	} else if strings.Contains(strings.ToLower(req.Model), "zhipu") || strings.Contains(strings.ToLower(req.Model), "glm") {
		providerName = "zhipu"
	}

	// If a specific provider was identified, try to use it
	if providerName != "" {
		client, exists := m.Providers[providerName]
		if exists {
			return client.ChatCompletion(ctx, req)
		}
	}

	// If no specific provider was found or the specific one doesn't exist, 
	// try to use any available provider
	for _, client := range m.Providers {
		// Just use the first available client as fallback
		return client.ChatCompletion(ctx, req)
	}

	return nil, fmt.Errorf("no AI provider available")
}