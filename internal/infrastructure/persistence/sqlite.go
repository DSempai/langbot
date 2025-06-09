package persistence

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		username TEXT,
		first_name TEXT,
		last_name TEXT,
		language_code TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// User preferences table for flexible settings
	userPreferencesTable := `
	CREATE TABLE IF NOT EXISTS user_preferences (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		preference_key TEXT NOT NULL,
		preference_value TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, preference_key)
	);`

	_, err = db.Exec(userPreferencesTable)
	if err != nil {
		return fmt.Errorf("failed to create user_preferences table: %w", err)
	}

	// Words table
	wordsTable := `
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		english TEXT NOT NULL,
		dutch TEXT NOT NULL,
		category TEXT NOT NULL,
		UNIQUE(english, dutch)
	);`

	_, err = db.Exec(wordsTable)
	if err != nil {
		return fmt.Errorf("failed to create words table: %w", err)
	}

	// User progress table with FSRS parameters
	userProgressTable := `
	CREATE TABLE IF NOT EXISTS user_progress (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word_id INTEGER NOT NULL,
		stability REAL DEFAULT 1.0,
		difficulty REAL DEFAULT 5.0,
		last_review DATETIME,
		due_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		review_count INTEGER DEFAULT 0,
		lapses INTEGER DEFAULT 0,
		state TEXT DEFAULT 'new',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (word_id) REFERENCES words (id),
		UNIQUE(user_id, word_id)
	);`

	_, err = db.Exec(userProgressTable)
	if err != nil {
		return fmt.Errorf("failed to create user_progress table: %w", err)
	}

	// Review history table
	reviewHistoryTable := `
	CREATE TABLE IF NOT EXISTS review_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word_id INTEGER NOT NULL,
		rating INTEGER NOT NULL,
		review_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		response_time_ms INTEGER,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (word_id) REFERENCES words (id)
	);`

	_, err = db.Exec(reviewHistoryTable)
	if err != nil {
		return fmt.Errorf("failed to create review_history table: %w", err)
	}

	// Drop and recreate grammar tips table with correct schema
	_, err = db.Exec("DROP TABLE IF EXISTS grammar_tips")
	if err != nil {
		return fmt.Errorf("failed to drop grammar_tips table: %w", err)
	}

	grammarTipsTable := `
	CREATE TABLE grammar_tips (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		explanation TEXT NOT NULL,
		dutch_example TEXT,
		english_example TEXT,
		category TEXT NOT NULL,
		applicable_categories TEXT DEFAULT '[]',
		word_patterns TEXT DEFAULT '[]',
		specific_words TEXT DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(title)
	);`

	_, err = db.Exec(grammarTipsTable)
	if err != nil {
		return fmt.Errorf("failed to create grammar_tips table: %w", err)
	}

	return nil
}
