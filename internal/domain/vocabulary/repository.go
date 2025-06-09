package vocabulary

import "context"

// Repository defines the contract for vocabulary persistence
type Repository interface {
	// Save persists a word to storage
	Save(ctx context.Context, word *Word) error

	// SaveBatch persists multiple words to storage
	SaveBatch(ctx context.Context, words []*Word) error

	// FindByID retrieves a word by its ID
	FindByID(ctx context.Context, id ID) (*Word, error)

	// FindAll retrieves all words
	FindAll(ctx context.Context) ([]*Word, error)

	// FindByCategory retrieves words by category
	FindByCategory(ctx context.Context, category Category) ([]*Word, error)

	// Exists checks if a word already exists
	Exists(ctx context.Context, english, dutch string) (bool, error)
}
