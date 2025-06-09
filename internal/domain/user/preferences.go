package user

import (
	"strconv"
	"time"
)

// Preference keys constants
const (
	PrefGrammarTipsEnabled    = "grammar_tips_enabled"
	PrefSmartRemindersEnabled = "smart_reminders_enabled"
)

// UserPreference represents a user preference
type UserPreference struct {
	id              ID
	userID          ID
	preferenceKey   string
	preferenceValue string
	createdAt       time.Time
	updatedAt       time.Time
}

// UserPreferences holds all user preferences
type UserPreferences struct {
	userID      ID
	preferences map[string]string
}

// NewUserPreferences creates a new user preferences with default values
func NewUserPreferences(userID ID) *UserPreferences {
	defaultPrefs := map[string]string{
		PrefGrammarTipsEnabled:    "true",
		PrefSmartRemindersEnabled: "true",
	}

	return &UserPreferences{
		userID:      userID,
		preferences: defaultPrefs,
	}
}

// UserPreferences methods
func (up *UserPreferences) UserID() ID {
	return up.userID
}

func (up *UserPreferences) GetBoolPreference(key string) bool {
	value, exists := up.preferences[key]
	if !exists {
		// Return default values for known preferences
		switch key {
		case PrefGrammarTipsEnabled, PrefSmartRemindersEnabled:
			return true
		default:
			return false
		}
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return boolValue
}

func (up *UserPreferences) SetBoolPreference(key string, value bool) {
	up.preferences[key] = strconv.FormatBool(value)
}

func (up *UserPreferences) GetStringPreference(key string) string {
	value, exists := up.preferences[key]
	if !exists {
		return ""
	}
	return value
}

func (up *UserPreferences) SetStringPreference(key, value string) {
	up.preferences[key] = value
}

func (up *UserPreferences) GetAllPreferences() map[string]string {
	return up.preferences
}

func (up *UserPreferences) SetPreferences(preferences map[string]string) {
	up.preferences = preferences
}

// Convenience methods for known preferences
func (up *UserPreferences) GrammarTipsEnabled() bool {
	return up.GetBoolPreference(PrefGrammarTipsEnabled)
}

func (up *UserPreferences) SetGrammarTipsEnabled(enabled bool) {
	up.SetBoolPreference(PrefGrammarTipsEnabled, enabled)
}

func (up *UserPreferences) ToggleGrammarTips() bool {
	newValue := !up.GrammarTipsEnabled()
	up.SetGrammarTipsEnabled(newValue)
	return newValue
}

func (up *UserPreferences) SmartRemindersEnabled() bool {
	return up.GetBoolPreference(PrefSmartRemindersEnabled)
}

func (up *UserPreferences) SetSmartRemindersEnabled(enabled bool) {
	up.SetBoolPreference(PrefSmartRemindersEnabled, enabled)
}

func (up *UserPreferences) ToggleSmartReminders() bool {
	newValue := !up.SmartRemindersEnabled()
	up.SetSmartRemindersEnabled(newValue)
	return newValue
}
