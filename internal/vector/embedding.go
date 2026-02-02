// Package vector provides text embedding and vector operations for OpenClaw-Go
package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// Embedding represents a text embedding vector
type Embedding struct {
	Values []float32  `json:"vector"`
	Model  string     `json:"model"`
}

// EmbedRequest represents a request to generate embeddings
type EmbedRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
}

// EmbedResponse represents the response from an embedding API
type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

// Embedder interface for text embedding generation
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	GetModelName() string
}

// OllamaEmbedder implements Embedder using Ollama's local API
type OllamaEmbedder struct {
	Endpoint string
	Model    string
	Client   *http.Client
}

// NewOllamaEmbedder creates a new Ollama-based embedder
func NewOllamaEmbedder(endpoint, model string) *OllamaEmbedder {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text" // Default embedding model for Ollama
	}

	return &OllamaEmbedder{
		Endpoint: endpoint,
		Model:    model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed generates an embedding for the given text
func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Truncate if too long (Ollama has context limits)
	maxTokens := 8192
	if len(text) > maxTokens*4 {
		text = text[:maxTokens*4]
	}

	reqBody := map[string]interface{}{
		"model":  o.Model,
		"prompt": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", 
		fmt.Sprintf("%s/api/embeddings", o.Endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (o *OllamaEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		emb, err := o.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	
	return embeddings, nil
}

// GetModelName returns the model name
func (o *OllamaEmbedder) GetModelName() string {
	return o.Model
}

// Similarity calculates cosine similarity between two vectors
func Similarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// DotProduct calculates dot product of two vectors
func DotProduct(a, b []float32) float32 {
	var result float32
	for i := range a {
		result += a[i] * b[i]
	}
	return result
}

// Normalize normalizes a vector to unit length
func Normalize(v []float32) []float32 {
	var norm float32
	for i := range v {
		norm += v[i] * v[i]
	}
	norm = float32(math.Sqrt(float64(norm)))
	
	if norm == 0 {
		return v
	}

	result := make([]float32, len(v))
	for i := range v {
		result[i] = v[i] / norm
	}
	return result
}
