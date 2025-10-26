package memory

import (
	"context"
	"fmt"

	"agent-workspace/backend/pkg/ollama"
)

// EmbeddingGenerator generates embeddings using Ollama
type EmbeddingGenerator struct {
	client *ollama.Client
}

// NewEmbeddingGenerator creates a new embedding generator
func NewEmbeddingGenerator() *EmbeddingGenerator {
	return &EmbeddingGenerator{
		client: ollama.NewClient(),
	}
}

// Generate generates an embedding for text
func (g *EmbeddingGenerator) Generate(ctx context.Context, text string) ([]float64, error) {
	return g.client.CreateEmbedding(text)
}

// GenerateBatch generates embeddings for multiple texts
func (g *EmbeddingGenerator) GenerateBatch(ctx context.Context, texts []string) ([][]float64, error) {
	return g.client.CreateEmbeddings(texts)
}

// CreateEmbeddingFunction creates an embedding function for LightRAG
func CreateEmbeddingFunction() func(context.Context, []string) ([][]float64, error) {
	generator := NewEmbeddingGenerator()
	
	return func(ctx context.Context, texts []string) ([][]float64, error) {
		return generator.GenerateBatch(ctx, texts)
	}
}

// CosineSimilarity calculates cosine similarity between two embeddings
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("embeddings must have same length")
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return dotProduct / (sqrt(normA) * sqrt(normB)), nil
}

// FindMostSimilar finds the most similar embedding from a list
func FindMostSimilar(query []float64, candidates [][]float64) (int, float64, error) {
	if len(candidates) == 0 {
		return -1, 0, fmt.Errorf("no candidates provided")
	}

	maxSimilarity := -1.0
	maxIndex := -1

	for i, candidate := range candidates {
		similarity, err := CosineSimilarity(query, candidate)
		if err != nil {
			continue
		}

		if similarity > maxSimilarity {
			maxSimilarity = similarity
			maxIndex = i
		}
	}

	if maxIndex == -1 {
		return -1, 0, fmt.Errorf("no valid similarities found")
	}

	return maxIndex, maxSimilarity, nil
}

// Helper function for square root
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	// Newton's method for square root
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}

	return z
}

