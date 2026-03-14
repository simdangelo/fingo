package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
    "time"
    "math/rand"
)

func IsValidHexColor(color string) bool {
    matched, _ := regexp.MatchString(`^#[0-9A-Fa-f]{6}$`, color)
    return matched
}

func generateID() string {
    return fmt.Sprintf("cat_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

type Category struct {
    ID string
	Name string
	Color string
}

func NewCategory(name string, color string) (*Category, error) {
    name = strings.TrimSpace(name)
    
    if name == "" {
        return nil, errors.New("category name cannot be empty")
    }
    if len(name) > 50 {
        return nil, errors.New("category name too long (max 50 characters)")
    }
    if !IsValidHexColor(color) {
        return nil, errors.New("invalid color format (expected #RRGGBB)")
    }
    
    return &Category{
        ID: generateID(),
        Name:  name,
        Color: color,
    }, nil
}
