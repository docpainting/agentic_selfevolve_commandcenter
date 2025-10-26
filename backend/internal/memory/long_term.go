package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	lightrag "github.com/MegaGrindStone/go-light-rag"
	"github.com/MegaGrindStone/go-light-rag/storage"
)

// LongTermMemory manages long-term knowledge storage
type LongTermMemory struct {
	rag          *lightrag.LightRAG
	neo4jStorage *storage.Neo4J
	chromemStore *storage.Chromem
	boltStore    *storage.Bolt
	mu           sync.RWMutex
	initialized  bool
}

// MemoryEntry represents a memory entry
type MemoryEntry struct {
	ID        string
	Type      string // "conversation", "code", "concept", "action"
	Content   string
	Metadata  map[string]interface{}
	Embedding []float64
	Timestamp time.Time
}

// NewLongTermMemory creates a new long-term memory system
func NewLongTermMemory() (*LongTermMemory, error) {
	// Initialize storage backends
	neo4j, err := initNeo4j()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Neo4j: %w", err)
	}

	chromem, err := initChromem()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ChromeM: %w", err)
	}

	bolt, err := initBolt()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Bolt: %w", err)
	}

	// Create embedding function
	embeddingFunc := createEmbeddingFunction()

	// Initialize LightRAG
	rag, err := lightrag.New(
		lightrag.WithGraphStorage(neo4j),
		lightrag.WithVectorStorage(chromem),
		lightrag.WithKVStorage(bolt),
		lightrag.WithEmbeddingFunc(embeddingFunc),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LightRAG: %w", err)
	}

	return &LongTermMemory{
		rag:          rag,
		neo4jStorage: neo4j,
		chromemStore: chromem,
		boltStore:    bolt,
		initialized:  true,
	}, nil
}

// Store stores content in long-term memory
func (m *LongTermMemory) Store(ctx context.Context, content string, metadata map[string]interface{}) error {
	if !m.initialized {
		return fmt.Errorf("memory system not initialized")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Insert into LightRAG
	return m.rag.Insert(ctx, content)
}

// Query queries the knowledge graph
func (m *LongTermMemory) Query(ctx context.Context, query string) (string, error) {
	if !m.initialized {
		return "", fmt.Errorf("memory system not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Query LightRAG
	result, err := m.rag.Query(ctx, query, lightrag.ModeHybrid)
	if err != nil {
		return "", fmt.Errorf("failed to query: %w", err)
	}

	return result, nil
}

// VectorSearch performs vector similarity search
func (m *LongTermMemory) VectorSearch(ctx context.Context, query string, topK int) ([]string, error) {
	if !m.initialized {
		return nil, fmt.Errorf("memory system not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Perform vector search through LightRAG
	result, err := m.rag.Query(ctx, query, lightrag.ModeLocal)
	if err != nil {
		return nil, fmt.Errorf("failed to vector search: %w", err)
	}

	return []string{result}, nil
}

// StoreConversation stores a conversation in memory
func (m *LongTermMemory) StoreConversation(ctx context.Context, userMsg, agentMsg string) error {
	content := fmt.Sprintf("User: %s\nAgent: %s", userMsg, agentMsg)
	metadata := map[string]interface{}{
		"type":      "conversation",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return m.Store(ctx, content, metadata)
}

// StoreCode stores code with context
func (m *LongTermMemory) StoreCode(ctx context.Context, filepath, code string, language string) error {
	content := fmt.Sprintf("File: %s\nLanguage: %s\nCode:\n%s", filepath, language, code)
	metadata := map[string]interface{}{
		"type":     "code",
		"filepath": filepath,
		"language": language,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return m.Store(ctx, content, metadata)
}

// StoreConcept stores a concept with relationships
func (m *LongTermMemory) StoreConcept(ctx context.Context, concept string, relationships map[string]string) error {
	content := fmt.Sprintf("Concept: %s\n", concept)
	for rel, target := range relationships {
		content += fmt.Sprintf("%s -> %s\n", rel, target)
	}

	metadata := map[string]interface{}{
		"type":          "concept",
		"concept":       concept,
		"relationships": relationships,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	return m.Store(ctx, content, metadata)
}

// StoreAction stores an action with result
func (m *LongTermMemory) StoreAction(ctx context.Context, action, result string, success bool) error {
	content := fmt.Sprintf("Action: %s\nResult: %s\nSuccess: %v", action, result, success)
	metadata := map[string]interface{}{
		"type":      "action",
		"action":    action,
		"success":   success,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return m.Store(ctx, content, metadata)
}

// GetContext retrieves relevant context for a query
func (m *LongTermMemory) GetContext(ctx context.Context, query string, maxTokens int) (string, error) {
	return m.Query(ctx, query)
}

// GetRelatedConcepts finds concepts related to a query
func (m *LongTermMemory) GetRelatedConcepts(ctx context.Context, concept string) ([]string, error) {
	query := fmt.Sprintf("Find concepts related to: %s", concept)
	result, err := m.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Parse result into concepts
	// TODO: Implement proper parsing
	return []string{result}, nil
}

// GetCodeContext retrieves code context
func (m *LongTermMemory) GetCodeContext(ctx context.Context, filepath string) (string, error) {
	query := fmt.Sprintf("Get code context for file: %s", filepath)
	return m.Query(ctx, query)
}

// Cleanup closes all connections
func (m *LongTermMemory) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	// Close storage backends
	if m.neo4jStorage != nil {
		// Neo4j close handled by driver
	}

	if m.chromemStore != nil {
		// ChromeM close
	}

	if m.boltStore != nil {
		// Bolt close
	}

	m.initialized = false
	return nil
}

// Helper functions

func initNeo4j() (*storage.Neo4J, error) {
	// Get configuration from environment
	uri := getEnv("NEO4J_URI", "bolt://localhost:7687")
	username := getEnv("NEO4J_USERNAME", "neo4j")
	password := getEnv("NEO4J_PASSWORD", "password")

	return storage.NewNeo4J(uri, username, password)
}

func initChromem() (*storage.Chromem, error) {
	// ChromeM uses in-memory storage by default
	dbPath := getEnv("CHROMEM_DB_PATH", "./data/chromem.db")
	
	// Create embedding function for ChromeM
	embeddingFunc := createEmbeddingFunction()
	
	return storage.NewChromem(dbPath, 5, embeddingFunc)
}

func initBolt() (*storage.Bolt, error) {
	dbPath := getEnv("BOLT_DB_PATH", "./data/bolt.db")
	return storage.NewBolt(dbPath)
}

func createEmbeddingFunction() func(context.Context, []string) ([][]float64, error) {
	return func(ctx context.Context, texts []string) ([][]float64, error) {
		// Use Ollama for embeddings
		embeddings := make([][]float64, len(texts))
		
		// TODO: Implement actual Ollama embedding calls
		// For now, return placeholder embeddings
		for i := range texts {
			// Create a simple embedding (should be replaced with actual Ollama call)
			embeddings[i] = make([]float64, 768) // Standard embedding size
			for j := range embeddings[i] {
				embeddings[i][j] = 0.1 // Placeholder value
			}
		}
		
		return embeddings, nil
	}
}

func getEnv(key, defaultValue string) string {
	// TODO: Implement environment variable reading
	// For now, return default value
	return defaultValue
}

