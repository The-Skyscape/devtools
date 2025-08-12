package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/The-Skyscape/devtools/pkg/ai"
)

// OllamaProvider implements the LLMProvider interface for Ollama
type OllamaProvider struct {
	host       string
	client     *http.Client
	configured bool
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(host string) *OllamaProvider {
	if host == "" {
		host = "http://localhost:11434"
	}
	return &OllamaProvider{
		host:       host,
		client:     &http.Client{Timeout: 60 * time.Second},
		configured: true,
	}
}

// Complete generates a completion
func (o *OllamaProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	// Convert messages to Ollama format
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
	}

	if req.Temperature > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %s", body)
	}

	var result struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		TotalDuration int64 `json:"total_duration"`
		LoadDuration  int64 `json:"load_duration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &ai.CompletionResponse{
		Content: result.Message.Content,
		Model:   req.Model,
	}, nil
}

// Stream generates a streaming completion
func (o *OllamaProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan string, error) {
	ch := make(chan string)
	
	go func() {
		defer close(ch)
		
		// Convert messages to Ollama format
		ollamaReq := map[string]interface{}{
			"model":    req.Model,
			"messages": req.Messages,
			"stream":   true,
		}

		if req.Temperature > 0 {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": req.Temperature,
			}
		}

		body, err := json.Marshal(ollamaReq)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/chat", bytes.NewReader(body))
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := o.client.Do(httpReq)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- fmt.Sprintf("Error: %s", body)
			return
		}

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}
			
			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				ch <- fmt.Sprintf("Error decoding: %v", err)
				break
			}
			
			if chunk.Message.Content != "" {
				ch <- chunk.Message.Content
			}
			
			if chunk.Done {
				break
			}
		}
	}()
	
	return ch, nil
}

// ListModels returns available models
func (o *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.host+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// IsConfigured returns whether the provider is configured
func (o *OllamaProvider) IsConfigured() bool {
	return o.configured && o.host != ""
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return "ollama"
}

// PullModel pulls a model from Ollama registry
func (o *OllamaProvider) PullModel(ctx context.Context, modelName string) error {
	req := map[string]interface{}{
		"name":   modelName,
		"stream": false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use longer timeout for model pulls
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to pull model: %s", body)
	}

	return nil
}

// Generate generates text (non-chat completion)
func (o *OllamaProvider) Generate(ctx context.Context, prompt, model string) (string, error) {
	req := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error: %s", body)
	}

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}