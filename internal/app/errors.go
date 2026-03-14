package app

import "errors"

var (
    // ErrCategoryNotFound indicates the specified category doesn't exist
    ErrCategoryNotFound = errors.New("category not found")
    
    // ErrTransactionNotFound indicates the specified transaction doesn't exist
    ErrTransactionNotFound = errors.New("transaction not found")
        
    // ErrCategoryInUse indicates a category cannot be deleted because transactions reference it
    ErrCategoryInUse = errors.New("category in use by transactions")
)