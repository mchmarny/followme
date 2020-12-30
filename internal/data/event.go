package data

const (
	// FollowedEventType when user followes
	FollowedEventType = "followed"

	// UnfollowedEventType when user unfollows
	UnfollowedEventType = "unfollowed"

	// FriendedEventType when user friends
	FriendedEventType = "friended"

	// UnfriendedEventType when user unfriends
	UnfriendedEventType = "unfriended"
)

// DayEvent represents day/label where day is not unique
type DayEvent struct {
	EventDate string `json:"event_date"`
	EventType string `json:"event_type"`
}

// GetFormattedEventType returns formatted event type
func (e *DayEvent) GetFormattedEventType() string {
	switch e.EventType {
	case FollowedEventType:
		return "they followed you"
	case UnfollowedEventType:
		return "they unfollowed you"
	case FriendedEventType:
		return "you followed them"
	case UnfriendedEventType:
		return "you unfollowed them"
	default:
		return e.EventType
	}
}

// UserEvent wraps simple twitter user as an time event
type UserEvent struct {
	*Profile
	EventDate   string `json:"event_at"`
	EventType   string `json:"event_type"`
	EventUser   string `json:"event_user"`
	IsFriend    bool   `json:"is_friend"`
	IsFollowing bool   `json:"is_following"`
}
