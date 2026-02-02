// Package ai provides AI model interfaces for OpenClaw-Go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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