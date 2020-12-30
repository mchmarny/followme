package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
)

var (
	listTypes = map[string]string{
		data.FollowedEventType:   "Who followed me",
		data.UnfollowedEventType: "Who unfollowed me",
		data.FriendedEventType:   "Whom I friended",
		data.UnfriendedEventType: "Whom I unfriended",
	}
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

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil || profile.Username == "" {
		a.viewErrorHandler(c, http.StatusBadRequest, err, "Error getting user profile")
		return
	}

	data := gin.H{
		"user":      profile,
		"version":   a.appVersion,
		"date":      isoDate,
		"listTypes": listTypes,
	}

	c.HTML(http.StatusOK, "day", data)
}
