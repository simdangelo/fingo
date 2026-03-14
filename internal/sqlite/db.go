// internal/sqlite/db.go
package sqlite

import (
    "database/sql"
    "fmt"
    
    _ "modernc.org/sqlite"
)

// Open opens a connection to the SQLite database at the given path.
// It enables foreign key constraints and verifies connectivity.
func Open(dbPath string) (*sql.DB, error) {
    // Open database (without connection string parameters)
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Verify connection
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Enable foreign keys using PRAGMA
    if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(1)
    
    return db, nil
}
// OpenInMemory creates an in-memory database for testing.
// The database is destroyed when the connection closes.
func OpenInMemory() (*sql.DB, error) {
    return Open(":memory:")
}

// InitDB opens the database and runs migrations.
// This is the main entry point for initializing the database.
func InitDB(dbPath string) (*sql.DB, error) {
    db, err := Open(dbPath)
    if err != nil {
        return nil, err
    }
    
    // Run migrations (implemented in next step)
    if err := Migrate(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("migration failed: %w", err)
    }
    
    return db, nil
}

// Close closes the database connection.
func Close(db *sql.DB) error {
    if db == nil {
        return nil
    }
    return db.Close()
}