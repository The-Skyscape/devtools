package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/The-Skyscape/devtools/pkg/ai"
)

// OllamaProvider implements the LLMProvider interface for Ollama
type OllamaProvider struct {
	host   string
	client *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(host string) *OllamaProvider {
	if host == "" {
		host = "http://localhost:11434"
	}
	
	return &OllamaProvider{
		host: strings.TrimSuffix(host, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Complete generates a completion for the given request
func (o *OllamaProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	// Prepare Ollama API request
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": o.messagesToPrompt(req.Messages),
		"stream": false,
		"options": map[string]interface{}{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
	}
	
	// Marshal request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", 
		fmt.Sprintf("%s/api/generate", o.host), 
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}
	
	// Parse response
	var ollamaResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract response content
	content := ""
	if response, ok := ollamaResp["response"].(string); ok {
		content = response
	}
	
	// Create completion response
	return &ai.CompletionResponse{
		Content:      content,
		Model:        req.Model,
		TokensUsed:   o.extractTokenCount(ollamaResp),
		FinishReason: "stop",
		Created:      time.Now(),
	}, nil
}

// Stream generates a streaming completion
func (o *OllamaProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan string, error) {
	ch := make(chan string)
	
	go func() {
		defer close(ch)
		
		// Prepare Ollama API request
		ollamaReq := map[string]interface{}{
			"model":  req.Model,
			"prompt": o.messagesToPrompt(req.Messages),
			"stream": true,
			"options": map[string]interface{}{
				"temperature": req.Temperature,
				"num_predict": req.MaxTokens,
			},
		}
		
		// Marshal request
		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		
		// Create HTTP request
		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			fmt.Sprintf("%s/api/generate", o.host),
			bytes.NewReader(reqBody))
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		
		httpReq.Header.Set("Content-Type", "application/json")
		
		// Send request
		resp, err := o.client.Do(httpReq)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		defer resp.Body.Close()
		
		// Read streaming response
		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				if err != io.EOF {
					ch <- fmt.Sprintf("Error: %v", err)
				}
				break
			}
			
			if response, ok := chunk["response"].(string); ok {
				ch <- response
			}
			
			if done, ok := chunk["done"].(bool); ok && done {
				break
			}
		}
	}()
	
	return ch, nil
}

// ListModels returns available models
func (o *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/tags", o.host), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Send request
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(body))
	}
	
	// Parse response
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract model names
	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}
	
	return models, nil
}

// IsConfigured checks if the provider is properly configured
func (o *OllamaProvider) IsConfigured() bool {
	// Check if Ollama is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/tags", o.host), nil)
	if err != nil {
		return false
	}
	
	resp, err := o.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return "ollama"
}

// Embed generates an embedding for the given text (if model supports it)
func (o *OllamaProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	// Prepare embedding request
	reqBody := map[string]interface{}{
		"model":  "llama2", // Default embedding model
		"prompt": text,
	}
	
	// Marshal request
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/embeddings", o.host),
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Send request
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", string(bodyBytes))
	}
	
	// Parse response
	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (o *OllamaProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	
	// Ollama doesn't support batch embeddings, so we do them one by one
	for i, text := range texts {
		embedding, err := o.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}
	
	return embeddings, nil
}

// Dimension returns the embedding dimension
func (o *OllamaProvider) Dimension() int {
	// This depends on the model, but common Ollama models use 4096
	return 4096
}

// messagesToPrompt converts messages to a single prompt string
func (o *OllamaProvider) messagesToPrompt(messages []ai.Message) string {
	var parts []string
	
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			parts = append(parts, fmt.Sprintf("System: %s", msg.Content))
		case "user":
			parts = append(parts, fmt.Sprintf("User: %s", msg.Content))
		case "assistant":
			parts = append(parts, fmt.Sprintf("Assistant: %s", msg.Content))
		}
	}
	
	// Add a prompt for the assistant to respond
	if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
		parts = append(parts, "Assistant:")
	}
	
	return strings.Join(parts, "\n\n")
}

// extractTokenCount extracts token count from Ollama response
func (o *OllamaProvider) extractTokenCount(resp map[string]interface{}) int {
	if evalCount, ok := resp["eval_count"].(float64); ok {
		return int(evalCount)
	}
	return 0
}