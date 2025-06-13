package user

import "context"

// Repository defines the contract for user persistence
type Repository interface {
	// Save persists a user to storage
	Save(ctx context.Context, user *User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id ID) (*User, error)

	// FindByTelegramID retrieves a user by their Telegram ID
	FindByTelegramID(ctx context.Context, telegramID TelegramID) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// UpdateLastActive updates the last active time of a user
	UpdateLastActive(ctx context.Context, id ID) error

	// GetAllUsers retrieves all users from storage
	GetAllUsers(ctx context.Context) ([]*User, error)
}
