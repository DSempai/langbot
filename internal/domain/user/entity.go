package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	id           ID
	telegramID   TelegramID
	username     string
	firstName    string
	lastName     string
	languageCode string
	createdAt    time.Time
	lastActive   time.Time
}

// ID represents the user's unique identifier
type ID int64

// TelegramID represents the user's Telegram ID
type TelegramID int64

// NewUser creates a new user
func NewUser(telegramID TelegramID, username, firstName, lastName, languageCode string) *User {
	now := time.Now()
	return &User{
		telegramID:   telegramID,
		username:     username,
		firstName:    firstName,
		lastName:     lastName,
		languageCode: languageCode,
		createdAt:    now,
		lastActive:   now,
	}
}

// Getters
func (u *User) ID() ID                 { return u.id }
func (u *User) TelegramID() TelegramID { return u.telegramID }
func (u *User) Username() string       { return u.username }
func (u *User) FirstName() string      { return u.firstName }
func (u *User) LastName() string       { return u.lastName }
func (u *User) LanguageCode() string   { return u.languageCode }
func (u *User) CreatedAt() time.Time   { return u.createdAt }
func (u *User) LastActive() time.Time  { return u.lastActive }

// SetID sets the user ID (used by repository)
func (u *User) SetID(id ID) {
	u.id = id
}

// UpdateLastActive updates the last active timestamp
func (u *User) UpdateLastActive() {
	u.lastActive = time.Now()
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(username, firstName, lastName, languageCode string) {
	u.username = username
	u.firstName = firstName
	u.lastName = lastName
	u.languageCode = languageCode
}
