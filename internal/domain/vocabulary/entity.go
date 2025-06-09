package vocabulary

// Word represents a vocabulary word with its translation
type Word struct {
	id       ID
	english  string
	dutch    string
	category Category
}

// ID represents the word's unique identifier
type ID int64

// Category represents the vocabulary category
type Category string

const (
	CategoryFamily     Category = "family"
	CategoryBody       Category = "body"
	CategoryColors     Category = "colors"
	CategoryFood       Category = "food"
	CategoryAnimals    Category = "animals"
	CategoryHome       Category = "home"
	CategoryObjects    Category = "objects"
	CategoryPeople     Category = "people"
	CategoryAdjectives Category = "adjectives"
	CategoryVerbs      Category = "verbs"
	CategoryParticles  Category = "particles"
)

// NewWord creates a new vocabulary word
func NewWord(english, dutch string, category Category) *Word {
	return &Word{
		english:  english,
		dutch:    dutch,
		category: category,
	}
}

// Getters
func (w *Word) ID() ID             { return w.id }
func (w *Word) English() string    { return w.english }
func (w *Word) Dutch() string      { return w.dutch }
func (w *Word) Category() Category { return w.category }

// SetID sets the word ID (used by repository)
func (w *Word) SetID(id ID) {
	w.id = id
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	switch Category(category) {
	case CategoryFamily, CategoryBody, CategoryColors, CategoryFood,
		CategoryAnimals, CategoryHome, CategoryObjects, CategoryPeople,
		CategoryAdjectives, CategoryVerbs, CategoryParticles:
		return true
	default:
		return false
	}
}
