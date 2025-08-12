package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Config represents the AI system configuration
type Config struct {
	// LLM settings
	DefaultProvider string
	DefaultModel    string
	Temperature     float64
	MaxTokens       int
	
	// Vector DB settings
	VectorDBProvider string
	
	// Rate limiting
	RequestsPerMinute int
	RequestsPerDay    int
	
	// Cache settings
	CacheTTL time.Duration
	
	// Provider-specific settings
	AnthropicAPIKey     string
	OpenAIAPIKey        string
	OllamaHost          string
	ChromaHost          string
	PineconeAPIKey      string
	PineconeEnvironment string
	WeaviateHost        string
	WeaviateAPIKey      string
}

// Manager orchestrates AI providers, vector databases, and tools
type Manager struct {
	config    *Config
	providers map[string]LLMProvider
	vectorDB  VectorDB
	embedder  EmbeddingProvider
	cache     Cache
	tools     map[string]Tool
	
	// Rate limiting
	requestCount int
	lastReset    time.Time
	mu           sync.RWMutex
}

// NewManager creates a new AI manager
func NewManager(config *Config) *Manager {
	m := &Manager{
		config:    config,
		providers: make(map[string]LLMProvider),
		tools:     make(map[string]Tool),
		lastReset: time.Now(),
	}
	
	// Initialize cache
	// m.cache = cache.NewMemoryCache() // Will be initialized when cache package is imported
	
	// Initialize providers based on config
	m.initializeProviders()
	
	return m
}

// initializeProviders sets up LLM and vector DB providers
func (m *Manager) initializeProviders() {
	// Initialize LLM providers
	if m.config.AnthropicAPIKey != "" {
		// Would initialize Anthropic provider here
		// m.providers["anthropic"] = llm.NewAnthropicProvider(m.config.AnthropicAPIKey)
	}
	
	if m.config.OpenAIAPIKey != "" {
		// Would initialize OpenAI provider here
		// m.providers["openai"] = llm.NewOpenAIProvider(m.config.OpenAIAPIKey)
	}
	
	if m.config.OllamaHost != "" {
		// Would initialize Ollama provider here
		// m.providers["ollama"] = llm.NewOllamaProvider(m.config.OllamaHost)
	}
	
	// Initialize vector DB
	switch m.config.VectorDBProvider {
	case "chroma":
		// m.vectorDB = vectordb.NewChromaDB(m.config.ChromaHost)
	case "pinecone":
		// m.vectorDB = vectordb.NewPinecone(m.config.PineconeAPIKey, m.config.PineconeEnvironment)
	case "weaviate":
		// m.vectorDB = vectordb.NewWeaviate(m.config.WeaviateHost, m.config.WeaviateAPIKey)
	}
	
	// Initialize embedding provider (usually from the default LLM provider)
	if provider, ok := m.providers[m.config.DefaultProvider]; ok {
		// Some providers can also generate embeddings
		if embedder, ok := provider.(EmbeddingProvider); ok {
			m.embedder = embedder
		}
	}
}

// GetProvider returns the specified LLM provider
func (m *Manager) GetProvider(name string) (LLMProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if name == "" {
		name = m.config.DefaultProvider
	}
	
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found or not configured", name)
	}
	
	return provider, nil
}

// GetDefaultProvider returns the default LLM provider
func (m *Manager) GetDefaultProvider() (LLMProvider, error) {
	return m.GetProvider(m.config.DefaultProvider)
}

// Complete generates a completion using the default provider
func (m *Manager) Complete(ctx context.Context, messages []Message) (*CompletionResponse, error) {
	// Check rate limiting
	if err := m.checkRateLimit(); err != nil {
		return nil, err
	}
	
	// Check cache
	cacheKey := m.getCacheKey(messages)
	if cached, ok := m.cache.Get(cacheKey); ok {
		if resp, ok := cached.(*CompletionResponse); ok {
			return resp, nil
		}
	}
	
	// Get provider
	provider, err := m.GetDefaultProvider()
	if err != nil {
		return nil, err
	}
	
	// Create request
	req := CompletionRequest{
		Messages:    messages,
		Model:       m.config.DefaultModel,
		Temperature: m.config.Temperature,
		MaxTokens:   m.config.MaxTokens,
	}
	
	// Generate completion
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// Cache response
	m.cache.Set(cacheKey, resp, m.config.CacheTTL)
	
	return resp, nil
}

// SearchSimilar performs vector similarity search
func (m *Manager) SearchSimilar(ctx context.Context, collection, query string, k int) ([]SearchResult, error) {
	if m.vectorDB == nil || !m.vectorDB.IsConfigured() {
		return nil, errors.New("vector database not configured")
	}
	
	if m.embedder == nil || !m.embedder.IsConfigured() {
		return nil, errors.New("embedding provider not configured")
	}
	
	// Generate embedding for query
	embedding, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	
	// Search in vector DB
	results, err := m.vectorDB.Search(ctx, collection, embedding, k)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	
	return results, nil
}

// IndexDocument adds a document to the vector database
func (m *Manager) IndexDocument(ctx context.Context, collection string, doc Document) error {
	if m.vectorDB == nil || !m.vectorDB.IsConfigured() {
		return errors.New("vector database not configured")
	}
	
	if m.embedder == nil || !m.embedder.IsConfigured() {
		return errors.New("embedding provider not configured")
	}
	
	// Generate embedding if not provided
	if len(doc.Vector) == 0 {
		embedding, err := m.embedder.Embed(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		doc.Vector = embedding
	}
	
	// Insert into vector DB
	if err := m.vectorDB.Insert(ctx, collection, []Document{doc}); err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	
	return nil
}

// RegisterTool registers a tool for use by agents
func (m *Manager) RegisterTool(tool Tool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools[tool.Name()] = tool
}

// GetTool returns a registered tool
func (m *Manager) GetTool(name string) (Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tool, ok := m.tools[name]
	return tool, ok
}

// GetAllTools returns all registered tools
func (m *Manager) GetAllTools() []Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tools := make([]Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}

// checkRateLimit checks if the rate limit has been exceeded
func (m *Manager) checkRateLimit() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Reset counter if needed
	if time.Since(m.lastReset) > time.Minute {
		m.requestCount = 0
		m.lastReset = time.Now()
	}
	
	// Check limit
	if m.config.RequestsPerMinute > 0 && m.requestCount >= m.config.RequestsPerMinute {
		return errors.New("rate limit exceeded")
	}
	
	m.requestCount++
	return nil
}

// getCacheKey generates a cache key for messages
func (m *Manager) getCacheKey(messages []Message) string {
	// Simple implementation - in production would use a proper hash
	key := fmt.Sprintf("%s:%s:", m.config.DefaultProvider, m.config.DefaultModel)
	for _, msg := range messages {
		key += fmt.Sprintf("%s:%s:", msg.Role, msg.Content)
	}
	return key
}

// IsConfigured checks if the AI system is properly configured
func (m *Manager) IsConfigured() bool {
	return len(m.providers) > 0 && m.config.DefaultProvider != ""
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// UpdateConfig updates the AI configuration
func (m *Manager) UpdateConfig(config *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.config = config
	m.providers = make(map[string]LLMProvider)
	m.initializeProviders()
}