package grammar

import "context"

// Repository defines the contract for grammar tips persistence
type Repository interface {
	// SaveBatch persists multiple grammar tips to storage
	SaveBatch(ctx context.Context, tips []*GrammarTip) error

	// FindApplicableToWord finds grammar tips that apply to a specific word
	FindApplicableToWord(ctx context.Context, dutchWord, englishWord, category string) ([]*GrammarTip, error)
}
