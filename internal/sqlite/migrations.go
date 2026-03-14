// internal/sqlite/migrations.go
package sqlite

import (
    "database/sql"
    "fmt"
)

// Migrate runs database migrations to set up the schema.
// It's safe to call multiple times (idempotent).
func Migrate(db *sql.DB) error {
    migrations := []string{
        createCategoriesTable,
        createTransactionsTable,
        createIndexes,
    }
    
    for i, migration := range migrations {
        if _, err := db.Exec(migration); err != nil {
            return fmt.Errorf("migration %d failed: %w", i, err)
        }
    }
    
    return nil
}

const createCategoriesTable = `
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL
);
`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    date TIMESTAMP NOT NULL,
    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('expense', 'income')),
    category_id TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_transactions_date 
ON transactions(date);

CREATE INDEX IF NOT EXISTS idx_transactions_category 
ON transactions(category_id);

CREATE INDEX IF NOT EXISTS idx_transactions_type 
ON transactions(type);
`