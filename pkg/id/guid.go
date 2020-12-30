package id

import "github.com/google/uuid"

// NewID generates new ID using UUID v4
func NewID() string {
	return uuid.New().String()
}
