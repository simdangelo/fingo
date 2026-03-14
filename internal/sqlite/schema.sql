-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL
);

-- Transactions table
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

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_transactions_date 
ON transactions(date);

CREATE INDEX IF NOT EXISTS idx_transactions_category 
ON transactions(category_id);

CREATE INDEX IF NOT EXISTS idx_transactions_type 
ON transactions(type);