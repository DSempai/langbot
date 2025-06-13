package persistence

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(1 * time.Minute) // Maximum idle time of a connection

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
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
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
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

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_preferences_user_key ON user_preferences(user_id, preference_key);",
		"CREATE INDEX IF NOT EXISTS idx_words_category ON words(category);",
		"CREATE INDEX IF NOT EXISTS idx_words_english ON words(english);",
		"CREATE INDEX IF NOT EXISTS idx_words_dutch ON words(dutch);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_user_id ON user_progress(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_word_id ON user_progress(word_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_due_date ON user_progress(due_date);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_user_due ON user_progress(user_id, due_date);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_state ON user_progress(state);",
		"CREATE INDEX IF NOT EXISTS idx_review_history_user_id ON review_history(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_review_history_word_id ON review_history(word_id);",
		"CREATE INDEX IF NOT EXISTS idx_review_history_user_word ON review_history(user_id, word_id);",
		"CREATE INDEX IF NOT EXISTS idx_grammar_tips_category ON grammar_tips(category);",
		// Add composite indexes for common query patterns
		"CREATE INDEX IF NOT EXISTS idx_user_progress_user_word ON user_progress(user_id, word_id);",
		"CREATE INDEX IF NOT EXISTS idx_review_history_user_time ON review_history(user_id, review_time);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_user_state ON user_progress(user_id, state);",
		"CREATE INDEX IF NOT EXISTS idx_user_progress_due_state ON user_progress(due_date, state);",
	}

	for _, idx := range indexes {
		_, err = db.Exec(idx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
