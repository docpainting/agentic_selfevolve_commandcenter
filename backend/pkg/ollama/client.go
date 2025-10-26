package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is an Ollama API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	model      string
	embedModel string
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a v1 chat completion request
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatCompletionResponse represents a v1 chat completion response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient creates a new Ollama client
func NewClient() *Client {
	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "gemma3:27b"
	}

	embedModel := os.Getenv("OLLAMA_EMBEDDING_MODEL")
	if embedModel == "" {
		embedModel = "nomic-embed-text:v1.5"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes for large models
		},
		model:      model,
		embedModel: embedModel,
	}
}

// ChatCompletion sends a chat completion request using v1 API
func (c *Client) ChatCompletion(messages []ChatMessage, temperature float64) (*ChatCompletionResponse, error) {
	req := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      false,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *Client) ChatCompletionStream(messages []ChatMessage, temperature float64, callback func(string) error) error {
	req := ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      true,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk ChatCompletionResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode chunk: %w", err)
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Message.Content
			if content != "" {
				if err := callback(content); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// CreateEmbedding creates an embedding for the given text
func (c *Client) CreateEmbedding(text string) ([]float64, error) {
	req := EmbeddingRequest{
		Model: c.embedModel,
		Input: text,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Data[0].Embedding, nil
}

// CreateEmbeddings creates embeddings for multiple texts
func (c *Client) CreateEmbeddings(texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embedding, err := c.CreateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

// GetModel returns the configured model name
func (c *Client) GetModel() string {
	return c.model
}

// GetEmbedModel returns the configured embedding model name
func (c *Client) GetEmbedModel() string {
	return c.embedModel
}

// Health checks if Ollama is accessible
func (c *Client) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("ollama not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

