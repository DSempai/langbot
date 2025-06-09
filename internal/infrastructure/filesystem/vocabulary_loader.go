package filesystem

import (
	"encoding/json"
	"fmt"
	"os"

	"dutch-learning-bot/internal/domain/vocabulary"
)

// VocabularyLoader handles loading vocabulary from files
type VocabularyLoader struct{}

// NewVocabularyLoader creates a new vocabulary loader
func NewVocabularyLoader() *VocabularyLoader {
	return &VocabularyLoader{}
}

// VocabularyData represents the JSON structure of vocabulary data
type VocabularyData struct {
	EnglishDutch []VocabularyEntry `json:"english_dutch"`
}

// VocabularyEntry represents a single vocabulary entry in JSON
type VocabularyEntry struct {
	Word        string `json:"word"`
	Translation string `json:"translation"`
	Category    string `json:"category"`
}

// LoadFromFile loads vocabulary from a JSON file
func (vl *VocabularyLoader) LoadFromFile(filename string) ([]*vocabulary.Word, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open vocabulary file: %w", err)
	}
	defer file.Close()

	var data VocabularyData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode vocabulary JSON: %w", err)
	}

	var words []*vocabulary.Word
	for _, entry := range data.EnglishDutch {
		// Validate category
		if !vocabulary.IsValidCategory(entry.Category) {
			return nil, fmt.Errorf("invalid category: %s", entry.Category)
		}

		word := vocabulary.NewWord(
			entry.Word,
			entry.Translation,
			vocabulary.Category(entry.Category),
		)
		words = append(words, word)
	}

	return words, nil
}
