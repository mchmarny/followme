package data

import (
	"fmt"
	"time"

	"github.com/mchmarny/followme/pkg/format"
)

// DailyState represents daily user state
type DailyState struct {
	// meta
	Key       string    `storm:"id" json:"jey"`
	Username  string    `json:"username"`
	StateOn   string    `json:"date"`
	UpdatedOn time.Time `json:"updated_on"`

	// follower
	Followers     []int64 `json:"followers"`
	FollowerCount int     `json:"follower_count"`

	NewFollowers     []int64 `json:"new_followers"`
	NewFollowerCount int     `json:"new_follower_count"`

	NewUnfollowers     []int64 `json:"new_unfollowers"`
	NewUnfollowerCount int     `json:"new_unfollower_count"`

	// friend
	Friends      []int64 `json:"friends"`
	FriendsCount int     `json:"friend_count"`

	NewFriends      []int64 `json:"new_friends"`
	NewFriendsCount int     `json:"new_friend_count"`

	NewUnfriended      []int64 `json:"new_unfriended"`
	NewUnfriendedCount int     `json:"new_unfriended_count"`
}

// GetDailyStateKey returns state key for a date
func GetDailyStateKey(username string, date time.Time) string {
	return fmt.Sprintf("%s-%s", format.NormalizeString(username), format.ToISODate(date))
}

// GetDailyStateKeyISO returns state key for an ISO date
func GetDailyStateKeyISO(username, isoDate string) string {
	return fmt.Sprintf("%s-%s", format.NormalizeString(username), isoDate)
}
