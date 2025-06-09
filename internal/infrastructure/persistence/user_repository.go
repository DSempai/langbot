package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dutch-learning-bot/internal/domain/user"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) user.Repository {
	return &userRepository{db: db}
}

// Save persists a user to storage
func (r *userRepository) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, language_code, created_at, last_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		int64(u.TelegramID()), u.Username(), u.FirstName(), u.LastName(),
		u.LanguageCode(), u.CreatedAt(), u.LastActive())
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	u.SetID(user.ID(id))
	return nil
}

// FindByID retrieves a user by their ID
func (r *userRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language_code, created_at, last_active
		FROM users WHERE id = ?
	`

	var telegramID int64
	var username, firstName, lastName, languageCode string
	var createdAt, lastActive time.Time

	err := r.db.QueryRowContext(ctx, query, int64(id)).Scan(
		&id, &telegramID, &username, &firstName, &lastName, &languageCode, &createdAt, &lastActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	u := user.NewUser(user.TelegramID(telegramID), username, firstName, lastName, languageCode)
	u.SetID(id)

	return u, nil
}

// FindByTelegramID retrieves a user by their Telegram ID
func (r *userRepository) FindByTelegramID(ctx context.Context, telegramID user.TelegramID) (*user.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, language_code, created_at, last_active
		FROM users WHERE telegram_id = ?
	`

	var id user.ID
	var tgID int64
	var username, firstName, lastName, languageCode string
	var createdAt, lastActive time.Time

	err := r.db.QueryRowContext(ctx, query, int64(telegramID)).Scan(
		&id, &tgID, &username, &firstName, &lastName, &languageCode, &createdAt, &lastActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by Telegram ID: %w", err)
	}

	u := user.NewUser(user.TelegramID(tgID), username, firstName, lastName, languageCode)
	u.SetID(id)

	return u, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users 
		SET username = ?, first_name = ?, last_name = ?, language_code = ?, last_active = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		u.Username(), u.FirstName(), u.LastName(), u.LanguageCode(), u.LastActive(), int64(u.ID()))
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
