package lightrag

import (
	"context"
	"log"
	"os"
)

// Example demonstrates how to use the LightRAG client
func Example() {
	ctx := context.Background()

	// Create logger
	logger := log.New(os.Stdout, "[LightRAG] ", log.LstdFlags)

	// Configure LightRAG
	cfg := &Config{
		// Neo4j configuration
		Neo4jURI:      "bolt://localhost:7687",
		Neo4jUser:     "neo4j",
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),

		// Vector DB configuration (ChromeM)
		VectorDBPath: "./data/vector.db",
		VectorDBSize: 10, // Return top 10 results

		// Key-Value DB configuration (BoltDB)
		KVDBPath: "./data/kv.db",

		// Ollama configuration
		OllamaBaseURL: "http://localhost:11434",
		LLMModel:      "gemma3:27b",
		EmbedModel:    "nomic-embed-text:v1.5",

		// Use default handler
		HandlerType: "default",

		Logger: logger,
	}

	// Create client
	client, err := NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create LightRAG client: %v", err)
	}
	defer client.Close()

	// Example 1: Insert Perception
	perceptionUUID, err := client.InsertPerception(
		ctx,
		"perception-001",
		"The HandleTask function is slow because it makes multiple database queries without caching.",
		map[string]interface{}{
			"type":       "perception",
			"confidence": 0.85,
		},
	)
	if err != nil {
		log.Fatalf("Failed to insert perception: %v", err)
	}
	logger.Printf("Inserted perception with UUID: %s", perceptionUUID)

	// Example 2: Insert Reasoning
	reasoningUUID, err := client.InsertReasoning(
		ctx,
		"reasoning-001",
		[]string{
			"Add caching layer to reduce database calls",
			"Optimize database queries with indexes",
			"Implement connection pooling",
		},
		"Add caching layer to reduce database calls",
		perceptionUUID,
	)
	if err != nil {
		log.Fatalf("Failed to insert reasoning: %v", err)
	}
	logger.Printf("Inserted reasoning with UUID: %s", reasoningUUID)

	// Example 3: Insert Action
	actionUUID, err := client.InsertAction(
		ctx,
		"action-001",
		"Implement Redis caching for database queries",
		"Successfully added caching layer. Performance improved by 40%.",
		reasoningUUID,
	)
	if err != nil {
		log.Fatalf("Failed to insert action: %v", err)
	}
	logger.Printf("Inserted action with UUID: %s", actionUUID)

	// Example 4: Insert Reflection
	reflectionUUID, err := client.InsertReflection(
		ctx,
		"reflection-001",
		[]string{
			"Caching significantly improves performance for read-heavy operations",
			"Redis is effective for caching database query results",
			"40% performance improvement validates the approach",
		},
		[]string{
			"Pattern: Database bottleneck → Add caching layer",
			"Pattern: Read-heavy operations → Use Redis cache",
		},
		actionUUID,
	)
	if err != nil {
		log.Fatalf("Failed to insert reflection: %v", err)
	}
	logger.Printf("Inserted reflection with UUID: %s", reflectionUUID)

	// Example 5: Query for similar situations
	result, err := client.Query(ctx, "How can I improve database performance?")
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	logger.Printf("Found %d local entities and %d global entities",
		len(result.LocalEntities), len(result.GlobalEntities))

	// Print relevant sources
	for i, source := range result.LocalSources {
		logger.Printf("Source %d (Relevance: %.2f):\n%s\n",
			i+1, source.Relevance, source.Content)
	}

	// Example 6: Retrieve by UUID
	source, err := client.QueryByUUID(ctx, perceptionUUID)
	if err != nil {
		log.Fatalf("Failed to retrieve by UUID: %v", err)
	}
	logger.Printf("Retrieved perception: %s", source.Content)
}

