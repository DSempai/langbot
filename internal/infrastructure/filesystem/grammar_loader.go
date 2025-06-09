package filesystem

import (
	"encoding/json"
	"fmt"
	"os"

	"dutch-learning-bot/internal/domain/grammar"
)

// GrammarLoader handles loading grammar tips from files
type GrammarLoader struct{}

// NewGrammarLoader creates a new grammar loader
func NewGrammarLoader() *GrammarLoader {
	return &GrammarLoader{}
}

// GrammarData represents the JSON structure of grammar tips data
type GrammarData struct {
	GrammarTips []GrammarTipEntry `json:"grammar_tips"`
}

// GrammarTipEntry represents a single grammar tip entry in JSON
type GrammarTipEntry struct {
	Title                string   `json:"title"`
	Explanation          string   `json:"explanation"`
	DutchExample         string   `json:"dutch_example"`
	EnglishExample       string   `json:"english_example"`
	Category             string   `json:"category"`
	ApplicableCategories []string `json:"applicable_categories"`
	WordPatterns         []string `json:"word_patterns"`
	SpecificWords        []string `json:"specific_words"`
}

// LoadFromFile loads grammar tips from a JSON file
func (gl *GrammarLoader) LoadFromFile(filename string) ([]*grammar.GrammarTip, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open grammar tips file: %w", err)
	}
	defer file.Close()

	var data GrammarData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode grammar tips JSON: %w", err)
	}

	var tips []*grammar.GrammarTip
	for _, entry := range data.GrammarTips {
		// Validate category
		if !grammar.IsValidCategory(grammar.Category(entry.Category)) {
			return nil, fmt.Errorf("invalid grammar category: %s", entry.Category)
		}

		tip := grammar.NewGrammarTip(
			entry.Title,
			entry.Explanation,
			entry.DutchExample,
			entry.EnglishExample,
			grammar.Category(entry.Category),
			entry.ApplicableCategories,
			entry.WordPatterns,
			entry.SpecificWords,
		)

		tips = append(tips, tip)
	}

	return tips, nil
}
