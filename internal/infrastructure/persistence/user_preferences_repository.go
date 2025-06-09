package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"dutch-learning-bot/internal/domain/user"
)

type userPreferencesRepository struct {
	db *sql.DB
}

// NewUserPreferencesRepository creates a new user preferences repository
func NewUserPreferencesRepository(db *sql.DB) user.PreferencesRepository {
	return &userPreferencesRepository{db: db}
}

// FindPreferences retrieves all preferences for a user
func (r *userPreferencesRepository) FindPreferences(ctx context.Context, userID user.ID) (*user.UserPreferences, error) {
	query := `
		SELECT preference_key, preference_value 
		FROM user_preferences 
		WHERE user_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to query user preferences: %w", err)
	}
	defer rows.Close()

	preferences := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan preference: %w", err)
		}
		preferences[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating preferences: %w", err)
	}

	userPrefs := user.NewUserPreferences(userID)
	if len(preferences) > 0 {
		userPrefs.SetPreferences(preferences)
	}

	return userPrefs, nil
}

// SavePreferences saves user preferences
func (r *userPreferencesRepository) SavePreferences(ctx context.Context, preferences *user.UserPreferences) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing preferences
	deleteQuery := `DELETE FROM user_preferences WHERE user_id = ?`
	_, err = tx.ExecContext(ctx, deleteQuery, int64(preferences.UserID()))
	if err != nil {
		return fmt.Errorf("failed to delete existing preferences: %w", err)
	}

	// Insert new preferences
	insertQuery := `
		INSERT INTO user_preferences (user_id, preference_key, preference_value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	for key, value := range preferences.GetAllPreferences() {
		_, err = tx.ExecContext(ctx, insertQuery, int64(preferences.UserID()), key, value)
		if err != nil {
			return fmt.Errorf("failed to insert preference %s: %w", key, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdatePreference updates a single preference
func (r *userPreferencesRepository) UpdatePreference(ctx context.Context, userID user.ID, key, value string) error {
	query := `
		INSERT INTO user_preferences (user_id, preference_key, preference_value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, preference_key) 
		DO UPDATE SET preference_value = excluded.preference_value, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, int64(userID), key, value)
	if err != nil {
		return fmt.Errorf("failed to update preference %s: %w", key, err)
	}

	return nil
}
