package data

import (
	"time"
)

// User represents authenticated user
type User struct {
	Username          string    `storm:"id" json:"username"`
	AccessTokenKey    string    `json:"access_token_key"`
	AccessTokenSecret string    `json:"access_token_secret"`
	UpdatedAt         time.Time `json:"updated_at"`
}
