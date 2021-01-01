package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
)

func (a *App) dayHandler(c *gin.Context) {
	forUser, err := a.getUser(c)
	if err != nil {
		a.logger.Printf("error getting user from context: %v", err)
		a.logOutHandler(c)
		return
	}

	isoDate := c.Param("day")
	if isoDate == "" {
		a.viewErrorHandler(c, http.StatusBadRequest, nil, "Day required (param: day)")
		return
	}

	var state data.DailyState
	stateKey := data.GetDailyStateKeyISO(forUser.Username, isoDate)
	if err := a.db.One("Key", stateKey, &state); err != nil {
		a.viewErrorHandler(c, http.StatusBadRequest, err, "error getting user state")
		return
	}

	listTypes := map[string]string{
		data.FollowedEventType:   fmt.Sprintf("Who followed me (%d)", state.NewFollowerCount),
		data.UnfollowedEventType: fmt.Sprintf("Who unfollowed me (%d)", state.NewUnfollowerCount),
		data.FriendedEventType:   fmt.Sprintf("Whom I friended (%d)", state.NewFriendsCount),
		data.UnfriendedEventType: fmt.Sprintf("Whom I unfriended (%d)", state.NewUnfriendedCount),
	}

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil || profile.Username == "" {
		a.viewErrorHandler(c, http.StatusBadRequest, err, "Error getting user profile")
		return
	}

	data := gin.H{
		"user":      profile,
		"version":   a.appVersion,
		"days":      isoDate,
		"listTypes": listTypes,
	}

	c.HTML(http.StatusOK, "day", data)
}
