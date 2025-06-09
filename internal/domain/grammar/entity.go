package grammar

import "time"

// GrammarTip represents a grammar learning tip
type GrammarTip struct {
	id                   ID
	title                string
	explanation          string
	dutchExample         string
	englishExample       string
	category             Category
	applicableCategories []string // Vocabulary categories this tip applies to
	wordPatterns         []string // Word patterns/endings this tip applies to
	specificWords        []string // Specific words this tip applies to
	createdAt            time.Time
}

// ID represents a grammar tip's unique identifier
type ID int64

// Category represents grammar tip categories
type Category string

const (
	CategoryArticles     Category = "articles"
	CategoryVerbs        Category = "verbs"
	CategoryWordOrder    Category = "word_order"
	CategoryPlurals      Category = "plurals"
	CategoryPronouns     Category = "pronouns"
	CategoryAdjectives   Category = "adjectives"
	CategoryPrepositions Category = "prepositions"
	CategoryGeneral      Category = "general"
)

// NewGrammarTip creates a new grammar tip
func NewGrammarTip(
	title, explanation, dutchExample, englishExample string,
	category Category,
	applicableCategories, wordPatterns, specificWords []string,
) *GrammarTip {
	return &GrammarTip{
		title:                title,
		explanation:          explanation,
		dutchExample:         dutchExample,
		englishExample:       englishExample,
		category:             category,
		applicableCategories: applicableCategories,
		wordPatterns:         wordPatterns,
		specificWords:        specificWords,
		createdAt:            time.Now(),
	}
}

// Getters
func (gt *GrammarTip) ID() ID                         { return gt.id }
func (gt *GrammarTip) Title() string                  { return gt.title }
func (gt *GrammarTip) Explanation() string            { return gt.explanation }
func (gt *GrammarTip) DutchExample() string           { return gt.dutchExample }
func (gt *GrammarTip) EnglishExample() string         { return gt.englishExample }
func (gt *GrammarTip) Category() Category             { return gt.category }
func (gt *GrammarTip) ApplicableCategories() []string { return gt.applicableCategories }
func (gt *GrammarTip) WordPatterns() []string         { return gt.wordPatterns }
func (gt *GrammarTip) SpecificWords() []string        { return gt.specificWords }
func (gt *GrammarTip) CreatedAt() time.Time           { return gt.createdAt }

// SetID sets the grammar tip ID (used by repository)
func (gt *GrammarTip) SetID(id ID) {
	gt.id = id
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category Category) bool {
	switch category {
	case CategoryArticles, CategoryVerbs, CategoryWordOrder, CategoryPlurals,
		CategoryPronouns, CategoryAdjectives, CategoryPrepositions, CategoryGeneral:
		return true
	default:
		return false
	}
}

// IsApplicableToWord checks if this tip applies to a specific word
func (gt *GrammarTip) IsApplicableToWord(dutchWord, englishWord, category string) bool {
	// Check specific words first
	for _, word := range gt.specificWords {
		if word == dutchWord || word == englishWord {
			return true
		}
	}

	// Check vocabulary categories
	for _, cat := range gt.applicableCategories {
		if cat == category {
			return true
		}
	}

	// Check word patterns (endings, etc.)
	for _, pattern := range gt.wordPatterns {
		if matchesPattern(dutchWord, pattern) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a word matches a pattern
func matchesPattern(word, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex later
	if len(pattern) == 0 {
		return false
	}

	// Check for suffix patterns (starting with -)
	if pattern[0] == '-' && len(pattern) > 1 {
		suffix := pattern[1:]
		return len(word) >= len(suffix) && word[len(word)-len(suffix):] == suffix
	}

	// Check for prefix patterns (ending with -)
	if len(pattern) > 1 && pattern[len(pattern)-1] == '-' {
		prefix := pattern[:len(pattern)-1]
		return len(word) >= len(prefix) && word[:len(prefix)] == prefix
	}

	// Exact match
	return word == pattern
}
