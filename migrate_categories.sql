-- Migration script to update category names in database
-- Run this BEFORE starting the application with updated vocabulary

BEGIN TRANSACTION;

-- Update particles to prepositions
UPDATE words SET category = 'prepositions' WHERE category = 'particles';
UPDATE user_progress SET category = 'prepositions' WHERE word_id IN (
    SELECT id FROM words WHERE category = 'prepositions'
);

-- Update verbs to verbs_infinitive (for "to be", "to have", etc.)
UPDATE words SET category = 'verbs_infinitive' WHERE category = 'verbs';

-- Update common_verbs to verbs_action (for "go", "come", etc.)
UPDATE words SET category = 'verbs_action' WHERE category = 'common_verbs';

-- Verify the migration
SELECT 'Words by category after migration:' as status;
SELECT category, COUNT(*) as word_count FROM words GROUP BY category ORDER BY word_count DESC;

COMMIT;

-- Note: Run this SQL script against your SQLite database:
-- sqlite3 dutch_learning.db < migrate_categories.sql 