package usecases

import (
	"context"
	"fmt"

	"dutch-learning-bot/internal/domain/user"
)

// UserUseCase handles user-related business operations
type UserUseCase struct {
	userRepo        user.Repository
	preferencesRepo user.PreferencesRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo user.Repository, preferencesRepo user.PreferencesRepository) *UserUseCase {
	return &UserUseCase{
		userRepo:        userRepo,
		preferencesRepo: preferencesRepo,
	}
}

// GetOrCreateUser gets an existing user or creates a new one
func (uc *UserUseCase) GetOrCreateUser(
	ctx context.Context,
	telegramID user.TelegramID,
	username, firstName, lastName, languageCode string,
) (*user.User, error) {
	// Try to find existing user
	existingUser, err := uc.userRepo.FindByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if existingUser != nil {
		// Update user activity and profile if needed
		existingUser.UpdateLastActive()
		existingUser.UpdateProfile(username, firstName, lastName, languageCode)

		err = uc.userRepo.Update(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// Create new user
	newUser := user.NewUser(telegramID, username, firstName, lastName, languageCode)
	err = uc.userRepo.Save(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to save new user: %w", err)
	}

	// Initialize default preferences for new user
	preferences := user.NewUserPreferences(newUser.ID())
	err = uc.preferencesRepo.SavePreferences(ctx, preferences)
	if err != nil {
		// Log error but don't fail user creation
		fmt.Printf("Warning: failed to initialize preferences for user %d: %v\n", newUser.ID(), err)
	}

	return newUser, nil
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, userID user.ID) (*user.User, error) {
	u, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

// UpdateUser updates a user's information
func (uc *UserUseCase) UpdateUser(ctx context.Context, user *user.User) error {
	err := uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUserPreferences retrieves user preferences
func (uc *UserUseCase) GetUserPreferences(ctx context.Context, userID user.ID) (*user.UserPreferences, error) {
	preferences, err := uc.preferencesRepo.FindPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return preferences, nil
}

// UpdateUserPreferences updates user preferences
func (uc *UserUseCase) UpdateUserPreferences(ctx context.Context, preferences *user.UserPreferences) error {
	err := uc.preferencesRepo.SavePreferences(ctx, preferences)
	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// ToggleGrammarTips toggles grammar tips preference for a user
func (uc *UserUseCase) ToggleGrammarTips(ctx context.Context, userID user.ID) (bool, error) {
	preferences, err := uc.GetUserPreferences(ctx, userID)
	if err != nil {
		return false, err
	}

	newState := preferences.ToggleGrammarTips()

	err = uc.UpdateUserPreferences(ctx, preferences)
	if err != nil {
		return false, err
	}

	return newState, nil
}

// ToggleSmartReminders toggles smart reminders preference for a user
func (uc *UserUseCase) ToggleSmartReminders(ctx context.Context, userID user.ID) (bool, error) {
	preferences, err := uc.GetUserPreferences(ctx, userID)
	if err != nil {
		return false, err
	}

	newState := preferences.ToggleSmartReminders()

	err = uc.UpdateUserPreferences(ctx, preferences)
	if err != nil {
		return false, err
	}

	return newState, nil
}
