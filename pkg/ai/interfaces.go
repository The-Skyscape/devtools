package ai

import (
	"context"
	"time"
)

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // The message content
}

// CompletionRequest represents a request for LLM completion
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

// CompletionResponse represents the LLM's response
type CompletionResponse struct {
	Content      string    `json:"content"`
	Model        string    `json:"model"`
	TokensUsed   int       `json:"tokens_used"`
	FinishReason string    `json:"finish_reason"`
	Created      time.Time `json:"created"`
}

// LLMProvider is the interface that all LLM providers must implement
type LLMProvider interface {
	// Complete generates a completion for the given request
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	
	// Stream generates a streaming completion
	Stream(ctx context.Context, req CompletionRequest) (<-chan string, error)
	
	// ListModels returns available models
	ListModels(ctx context.Context) ([]string, error)
	
	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
	
	// Name returns the provider name
	Name() string
}

// Embedding represents a vector embedding
type Embedding struct {
	Vector   []float64 `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Document represents a document for vector storage
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float64              `json:"vector,omitempty"`
}

// SearchResult represents a vector search result
type SearchResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
}

// VectorDB is the interface for vector databases
type VectorDB interface {
	// CreateCollection creates a new collection
	CreateCollection(ctx context.Context, name string, dimension int) error
	
	// DeleteCollection deletes a collection
	DeleteCollection(ctx context.Context, name string) error
	
	// Insert adds documents to a collection
	Insert(ctx context.Context, collection string, documents []Document) error
	
	// Search performs similarity search
	Search(ctx context.Context, collection string, query []float64, k int) ([]SearchResult, error)
	
	// Delete removes documents by ID
	Delete(ctx context.Context, collection string, ids []string) error
	
	// IsConfigured checks if the vector DB is properly configured
	IsConfigured() bool
	
	// Name returns the provider name
	Name() string
}

// Tool represents an AI tool that can be used by agents
type Tool interface {
	// Name returns the tool name
	Name() string
	
	// Description returns a description of what the tool does
	Description() string
	
	// Parameters returns the tool's parameter schema
	Parameters() map[string]interface{}
	
	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Agent represents an AI agent that can use tools and complete tasks
type Agent interface {
	// Name returns the agent name
	Name() string
	
	// Description returns the agent's capabilities
	Description() string
	
	// SetTools configures the tools available to the agent
	SetTools(tools []Tool)
	
	// Execute runs the agent with the given task
	Execute(ctx context.Context, task string) (*AgentResponse, error)
	
	// Chat handles a conversational interaction
	Chat(ctx context.Context, messages []Message) (*AgentResponse, error)
}

// AgentResponse represents an agent's response
type AgentResponse struct {
	Content      string                   `json:"content"`
	ToolCalls    []ToolCall               `json:"tool_calls,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
	TokensUsed   int                      `json:"tokens_used"`
}

// ToolCall represents a tool invocation by an agent
type ToolCall struct {
	Tool       string                 `json:"tool"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     interface{}            `json:"result,omitempty"`
}

// Cache interface for caching AI responses
type Cache interface {
	// Get retrieves a cached value
	Get(key string) (interface{}, bool)
	
	// Set stores a value in the cache
	Set(key string, value interface{}, ttl time.Duration)
	
	// Delete removes a value from the cache
	Delete(key string)
	
	// Clear removes all values from the cache
	Clear()
}

// EmbeddingProvider generates embeddings for text
type EmbeddingProvider interface {
	// Embed generates an embedding for the given text
	Embed(ctx context.Context, text string) ([]float64, error)
	
	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
	
	// Dimension returns the embedding dimension
	Dimension() int
	
	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
	
	// Name returns the provider name
	Name() string
}