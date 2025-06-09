package user

import "context"

// PreferencesRepository handles user preferences persistence
type PreferencesRepository interface {
	// FindPreferences retrieves all preferences for a user
	FindPreferences(ctx context.Context, userID ID) (*UserPreferences, error)

	// SavePreferences saves user preferences
	SavePreferences(ctx context.Context, preferences *UserPreferences) error

	// UpdatePreference updates a single preference
	UpdatePreference(ctx context.Context, userID ID, key, value string) error
}
