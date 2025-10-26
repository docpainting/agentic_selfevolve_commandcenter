package lightrag

import (
	"context"
	"fmt"
	"log"

	golightrag "github.com/MegaGrindStone/go-light-rag"
	"github.com/MegaGrindStone/go-light-rag/handler"
	"github.com/MegaGrindStone/go-light-rag/llm"
	"github.com/MegaGrindStone/go-light-rag/storage"
	"github.com/philippgille/chromem-go"
)

// Client wraps go-light-rag with our application-specific logic
type Client struct {
	llm     llm.LLM
	store   Storage
	handler handler.DocumentHandler
	logger  *log.Logger
}

// Storage combines all storage interfaces
type Storage struct {
	Graph  storage.GraphStorage
	Vector storage.VectorStorage
	KV     storage.KeyValueStorage
}

// Config holds configuration for LightRAG client
type Config struct {
	// Neo4j configuration
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string

	// Vector DB configuration (ChromeM)
	VectorDBPath string
	VectorDBSize int // Max results to return

	// Key-Value DB configuration (BoltDB)
	KVDBPath string

	// Ollama configuration
	OllamaBaseURL string
	LLMModel      string // e.g., "gemma3:27b"
	EmbedModel    string // e.g., "nomic-embed-text:v1.5"

	// Handler type
	HandlerType string // "default", "semantic", "go"

	Logger *log.Logger
}

// NewClient creates a new LightRAG client
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	// Initialize LLM (Ollama with Gemma 3)
	llmClient := llm.NewOllama(
		cfg.OllamaBaseURL,
		cfg.LLMModel,
		nil, // Use default params
		cfg.Logger,
	)

	// Initialize Neo4j for graph storage
	graphDB, err := storage.NewNeo4J(
		cfg.Neo4jURI,
		cfg.Neo4jUser,
		cfg.Neo4jPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Neo4j: %w", err)
	}

	// Create embedding function using Ollama
	embeddingFunc := func(ctx context.Context, text string) ([]float32, error) {
		// Call Ollama embeddings API
		// Note: This is a simplified version - you may need to implement
		// the actual Ollama embeddings API call
		return ollamaEmbedding(cfg.OllamaBaseURL, cfg.EmbedModel, text)
	}

	// Initialize ChromeM for vector storage
	vecDB, err := storage.NewChromem(
		cfg.VectorDBPath,
		cfg.VectorDBSize,
		storage.EmbeddingFunc(embeddingFunc),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ChromeM: %w", err)
	}

	// Initialize BoltDB for key-value storage
	kvDB, err := storage.NewBolt(cfg.KVDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize BoltDB: %w", err)
	}

	store := Storage{
		Graph:  graphDB,
		Vector: vecDB,
		KV:     kvDB,
	}

	// Select handler based on configuration
	var docHandler handler.DocumentHandler
	switch cfg.HandlerType {
	case "semantic":
		docHandler = &handler.Semantic{}
	case "go":
		docHandler = &handler.Go{}
	default:
		docHandler = &handler.Default{}
	}

	return &Client{
		llm:     llmClient,
		store:   store,
		handler: docHandler,
		logger:  cfg.Logger,
	}, nil
}

// InsertPerception stores perception data in LightRAG
func (c *Client) InsertPerception(ctx context.Context, id string, content string, metadata map[string]interface{}) (string, error) {
	doc := golightrag.Document{
		ID:      id,
		Content: content,
	}

	err := golightrag.Insert(doc, c.handler, c.store, c.llm, c.logger)
	if err != nil {
		return "", fmt.Errorf("failed to insert perception: %w", err)
	}

	// Return the document ID as UUID
	return id, nil
}

// InsertReasoning stores reasoning data in LightRAG
func (c *Client) InsertReasoning(ctx context.Context, id string, branches []string, selected string, perceptionUUID string) (string, error) {
	// Combine branches into content
	content := fmt.Sprintf("Reasoning Branches:\n")
	for i, branch := range branches {
		content += fmt.Sprintf("Branch %d: %s\n", i+1, branch)
	}
	content += fmt.Sprintf("\nSelected: %s\n", selected)
	content += fmt.Sprintf("Based on Perception: %s\n", perceptionUUID)

	doc := golightrag.Document{
		ID:      id,
		Content: content,
	}

	err := golightrag.Insert(doc, c.handler, c.store, c.llm, c.logger)
	if err != nil {
		return "", fmt.Errorf("failed to insert reasoning: %w", err)
	}

	return id, nil
}

// InsertAction stores action execution data in LightRAG
func (c *Client) InsertAction(ctx context.Context, id string, plan string, result string, reasoningUUID string) (string, error) {
	content := fmt.Sprintf("Action Plan:\n%s\n\nResult:\n%s\n\nBased on Reasoning: %s\n", plan, result, reasoningUUID)

	doc := golightrag.Document{
		ID:      id,
		Content: content,
	}

	err := golightrag.Insert(doc, c.handler, c.store, c.llm, c.logger)
	if err != nil {
		return "", fmt.Errorf("failed to insert action: %w", err)
	}

	return id, nil
}

// InsertReflection stores reflection data in LightRAG
func (c *Client) InsertReflection(ctx context.Context, id string, learnings []string, patterns []string, actionUUID string) (string, error) {
	content := fmt.Sprintf("Learnings:\n")
	for i, learning := range learnings {
		content += fmt.Sprintf("%d. %s\n", i+1, learning)
	}

	content += fmt.Sprintf("\nPatterns Discovered:\n")
	for i, pattern := range patterns {
		content += fmt.Sprintf("%d. %s\n", i+1, pattern)
	}

	content += fmt.Sprintf("\nBased on Action: %s\n", actionUUID)

	doc := golightrag.Document{
		ID:      id,
		Content: content,
	}

	err := golightrag.Insert(doc, c.handler, c.store, c.llm, c.logger)
	if err != nil {
		return "", fmt.Errorf("failed to insert reflection: %w", err)
	}

	return id, nil
}

// Query performs semantic search across all stored data
func (c *Client) Query(ctx context.Context, query string) (*QueryResult, error) {
	conversation := []golightrag.QueryConversation{
		{
			Role:    golightrag.RoleUser,
			Message: query,
		},
	}

	result, err := golightrag.Query(conversation, c.handler, c.store, c.llm, c.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	// Convert to our result format
	queryResult := &QueryResult{
		LocalEntities:  make([]Entity, len(result.LocalEntities)),
		GlobalEntities: make([]Entity, len(result.GlobalEntities)),
		LocalSources:   make([]Source, len(result.LocalSources)),
		GlobalSources:  make([]Source, len(result.GlobalSources)),
	}

	for i, entity := range result.LocalEntities {
		queryResult.LocalEntities[i] = Entity{
			Name:        entity.Name,
			Type:        entity.Type,
			Description: entity.Description,
		}
	}

	for i, entity := range result.GlobalEntities {
		queryResult.GlobalEntities[i] = Entity{
			Name:        entity.Name,
			Type:        entity.Type,
			Description: entity.Description,
		}
	}

	for i, source := range result.LocalSources {
		queryResult.LocalSources[i] = Source{
			ID:        source.ID,
			Content:   source.Content,
			Relevance: source.Relevance,
		}
	}

	for i, source := range result.GlobalSources {
		queryResult.GlobalSources[i] = Source{
			ID:        source.ID,
			Content:   source.Content,
			Relevance: source.Relevance,
		}
	}

	return queryResult, nil
}

// QueryByUUID retrieves data by UUID
func (c *Client) QueryByUUID(ctx context.Context, uuid string) (*Source, error) {
	// Query using the UUID as the search term
	result, err := c.Query(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Return the most relevant source
	if len(result.LocalSources) > 0 {
		return &result.LocalSources[0], nil
	}

	if len(result.GlobalSources) > 0 {
		return &result.GlobalSources[0], nil
	}

	return nil, fmt.Errorf("no data found for UUID: %s", uuid)
}

// Close closes all storage connections
func (c *Client) Close() error {
	// Close all storage backends
	// Note: Implement Close methods as needed
	return nil
}

// QueryResult holds the results from a LightRAG query
type QueryResult struct {
	LocalEntities  []Entity
	GlobalEntities []Entity
	LocalSources   []Source
	GlobalSources  []Source
}

// Entity represents a knowledge graph entity
type Entity struct {
	Name        string
	Type        string
	Description string
}

// Source represents a source document
type Source struct {
	ID        string
	Content   string
	Relevance float64
}

// ollamaEmbedding calls Ollama embeddings API
func ollamaEmbedding(baseURL, model, text string) ([]float32, error) {
	// TODO: Implement actual Ollama embeddings API call
	// This is a placeholder - you'll need to implement the HTTP call to Ollama
	// Example: POST to http://localhost:11434/api/embeddings
	// with body: {"model": "nomic-embed-text:v1.5", "prompt": text}
	
	// For now, return error to indicate not implemented
	return nil, fmt.Errorf("Ollama embedding not yet implemented - use chromem.NewEmbeddingFuncOllama instead")
}

