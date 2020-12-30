package data

import (
	"time"

	"github.com/mchmarny/followme/pkg/format"
)

// Profile represents simplified Twitter user profile
type Profile struct {
	ID            int64     `storm:"id" json:"id"`
	Username      string    `storm:"unique" json:"username"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ProfileImage  string    `json:"profile_image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Lang          string    `json:"lang"`
	Location      string    `json:"location"`
	Timezone      string    `json:"time_zone"`
	PostCount     int       `json:"post_count"`
	FaveCount     int       `json:"fave_count"`
	FriendCount   int       `json:"friend_count"`
	FollowerCount int       `json:"followers_count"`
	ListedCount   int       `json:"listed_count"`
}

// HasName is a template helper
func (p *Profile) HasName() bool {
	return p.Name != ""
}

// FormattedCreatedAt returns RFC822 formatted CreatedAt
func (p *Profile) FormattedCreatedAt() string {
	if p.CreatedAt.IsZero() {
		return ""
	}
	return p.CreatedAt.Format(time.RFC822)
}

// FormattedUpdatedAt returns RFC822 formatted UpdatedAt
func (p *Profile) FormattedUpdatedAt() string {
	if p.UpdatedAt.IsZero() {
		return ""
	}
	return p.UpdatedAt.Format(time.RFC822)
}

// UserSince displays the length of time since the user joined Twitter s
func (p *Profile) UserSince() string {
	if p.CreatedAt.IsZero() {
		return ""
	}
	return format.PrettyDurationSince(p.CreatedAt)
}
