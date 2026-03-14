package sqlite

import (
    "testing"
)

func TestMigrate(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    // Run migrations
    if err := Migrate(db); err != nil {
        t.Fatalf("Migrate() error = %v", err)
    }
    
    // Verify categories table exists
    var tableName string
    err = db.QueryRow(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name='categories'
    `).Scan(&tableName)
    
    if err != nil {
        t.Fatalf("categories table not found: %v", err)
    }
    
    if tableName != "categories" {
        t.Errorf("expected table name 'categories', got '%s'", tableName)
    }
}

func TestMigrate_Idempotent(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    // Run migrations multiple times
    for i := 0; i < 3; i++ {
        if err := Migrate(db); err != nil {
            t.Errorf("Migrate() run %d error = %v", i+1, err)
        }
    }
}

func TestMigrate_AllTables(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    if err := Migrate(db); err != nil {
        t.Fatalf("Migrate() error = %v", err)
    }
    
    // Check all tables exist
    expectedTables := []string{"categories", "transactions"}
    
    for _, tableName := range expectedTables {
        var name string
        err = db.QueryRow(`
            SELECT name FROM sqlite_master 
            WHERE type='table' AND name=?
        `, tableName).Scan(&name)
        
        if err != nil {
            t.Errorf("table %s not found: %v", tableName, err)
        }
    }
}

func TestMigrate_Indexes(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    if err := Migrate(db); err != nil {
        t.Fatalf("Migrate() error = %v", err)
    }
    
    // Check indexes exist
    expectedIndexes := []string{
        "idx_transactions_date",
        "idx_transactions_category",
        "idx_transactions_type",
    }
    
    for _, indexName := range expectedIndexes {
        var name string
        err = db.QueryRow(`
            SELECT name FROM sqlite_master 
            WHERE type='index' AND name=?
        `, indexName).Scan(&name)
        
        if err != nil {
            t.Errorf("index %s not found: %v", indexName, err)
        }
    }
}

func TestMigrate_ForeignKeyConstraint(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    if err := Migrate(db); err != nil {
        t.Fatalf("Migrate() error = %v", err)
    }
    
    // Try to insert transaction with invalid category_id
    _, err = db.Exec(`
        INSERT INTO transactions 
        (id, date, amount, type, category_id, description, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, "tx1", "2026-01-01T00:00:00Z", 50.0, "expense", "invalid_cat", "test", "2026-01-01T00:00:00Z")
    
    // Should fail due to foreign key constraint
    if err == nil {
        t.Error("expected foreign key constraint error, got nil")
    }
}

func TestMigrate_CheckConstraint(t *testing.T) {
    db, err := OpenInMemory()
    if err != nil {
        t.Fatalf("OpenInMemory() error = %v", err)
    }
    defer Close(db)
    
    if err := Migrate(db); err != nil {
        t.Fatalf("Migrate() error = %v", err)
    }
    
    // Insert valid category first
    _, err = db.Exec(`
        INSERT INTO categories (id, name, color)
        VALUES (?, ?, ?)
    `, "cat1", "Test", "#FF5733")
    
    if err != nil {
        t.Fatalf("failed to insert category: %v", err)
    }
    
    // Try to insert transaction with invalid type
    _, err = db.Exec(`
        INSERT INTO transactions 
        (id, date, amount, type, category_id, description, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, "tx1", "2026-01-01T00:00:00Z", 50.0, "invalid_type", "cat1", "test", "2026-01-01T00:00:00Z")
    
    // Should fail due to CHECK constraint
    if err == nil {
        t.Error("expected CHECK constraint error, got nil")
    }
}