package sqlite

import (
    "testing"
)

func TestOpen(t *testing.T) {
    // Use in-memory database for testing
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    // Verify connection works
    if err := db.Ping(); err != nil {
        t.Errorf("Ping() error = %v", err)
    }
}

func TestOpen_FileDatabase(t *testing.T) {
    // Create temporary database
    db, err := Open(t.TempDir() + "/test.db")
    if err != nil {
        t.Fatalf("Open() error = %v", err)
    }
    defer Close(db)
    
    // Verify connection
    if err := db.Ping(); err != nil {
        t.Errorf("Ping() error = %v", err)
    }
}

func TestForeignKeysEnabled(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    // Check if foreign keys are enabled
    var enabled int
    err = db.QueryRow("PRAGMA foreign_keys").Scan(&enabled)
    if err != nil {
        t.Fatalf("Query error = %v", err)
    }
    
    if enabled != 1 {
        t.Errorf("Foreign keys not enabled: got %v, want 1", enabled)
    }
}

func TestInitDB(t *testing.T) {
    // Create temporary directory
    tmpDir := t.TempDir()
    dbPath := tmpDir + "/test.db"
    
    // Initialize database
    db, err := InitDB(dbPath)
    if err != nil {
        t.Fatalf("InitDB() error = %v", err)
    }
    defer Close(db)
    
    // Verify tables were created
    var tableName string
    err = db.QueryRow(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name='categories'
    `).Scan(&tableName)
    
    if err != nil {
        t.Errorf("categories table not created: %v", err)
    }
}