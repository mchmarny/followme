package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
)

func (a *App) dayHandler(c *gin.Context) {
	ctx := c.Request.Context()
	forUser, err := a.getUser(c)
	if err != nil {
		a.logger.Printf("error getting user from context: %v", err)
		a.logOutHandler(c)
		return
	}

	profile, err := a.getUserProfile(c)
	if err != nil {
		a.logger.Printf("error getting user profile from context: %v", err)
		a.logOutHandler(c)
		return
	}

	isoDate := c.Param("day")
	if isoDate == "" {
		a.viewErrorHandler(c, http.StatusBadRequest, nil, "Day required (param: day)")
		return
	}

	var dayState data.DailyState
	stateKey := data.GetDailyStateKey(forUser.Username, time.Now().UTC())
	if err := a.db.One("Key", stateKey, &dayState); err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting daily follower state")
		return
	}

	newFollowerIDs := dayState.NewFollowers
	var newFollowersLimit bool
	if dayState.NewFollowerCount > maxEventLimit {
		newFollowerIDs = newFollowerIDs[len(newFollowerIDs)-maxEventLimit:]
		newFollowersLimit = true
	}

	followers, err := a.toUserEvent(ctx, forUser, dayState.Friends, newFollowerIDs, isoDate, data.FollowedEventType)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting new follower events")
		return
	}
	a.logger.Printf("Followers:%d, expected:%d", len(followers), dayState.FollowerCount)

	newUnfollowerIDs := dayState.NewUnfollowers
	var newUnfollowersLimit bool
	if dayState.NewUnfollowerCount > maxEventLimit {
		newUnfollowerIDs = newUnfollowerIDs[len(newUnfollowerIDs)-maxEventLimit:]
		newUnfollowersLimit = true
	}

	unfollowers, err := a.toUserEvent(ctx, forUser, dayState.Friends, newUnfollowerIDs, isoDate, data.UnfollowedEventType)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting unfollower events")
		return
	}
	a.logger.Printf("Unfollowers:%d, expected:%d", len(unfollowers), dayState.NewUnfollowerCount)

	newFriendedIDs := dayState.NewFriends
	var newFriendLimited bool
	if dayState.NewFriendsCount > maxEventLimit {
		newFriendedIDs = newFriendedIDs[len(newFriendedIDs)-maxEventLimit:]
		newFriendLimited = true
	}

	friended, err := a.toUserEvent(ctx, forUser, dayState.Friends, newFriendedIDs, isoDate, data.FriendedEventType)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting friended events")
		return
	}
	a.logger.Printf("Friends:%d, expected:%d", len(friended), dayState.NewFriendsCount)

	newUnfriendedIDs := dayState.NewUnfriended
	var newUnfriendLimited bool
	if dayState.NewUnfriendedCount > maxEventLimit {
		newUnfriendedIDs = newUnfriendedIDs[len(newUnfriendedIDs)-maxEventLimit:]
		newUnfriendLimited = true
	}

	unfriended, err := a.toUserEvent(ctx, forUser, dayState.Friends, newUnfriendedIDs, isoDate, data.UnfriendedEventType)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting unfriended events")
		return
	}
	a.logger.Printf("Friends:%d, expected:%d", len(unfriended), dayState.NewUnfriendedCount)

	data := gin.H{
		"user":       profile,
		"version":    a.appVersion,
		"date":       isoDate,
		"state":      dayState,
		"eventLimit": maxEventLimit,

		"followers":        followers,
		"hasFollowers":     len(followers) > 0,
		"followersLimited": newFollowersLimit,

		"unfollowers":        unfollowers,
		"hasUnfollowers":     len(unfollowers) > 0,
		"unfollowersLimited": newUnfollowersLimit,

		"friended":        friended,
		"hasFriended":     len(friended) > 0,
		"friendedLimited": newFriendLimited,

		"unfriended":        unfriended,
		"hasUnfriended":     len(unfriended) > 0,
		"unfriendedLimited": newUnfriendLimited,
	}

	c.HTML(http.StatusOK, "day", data)
}

func (a *App) toUserEvent(ctx context.Context, forUser *data.User, friends, ids []int64, isoDate, eventType string) (list []*data.UserEvent, err error) {
	if len(ids) > 0 {
		users, detailErr := a.twClient.GetUserDetailsFromIDs(ctx, forUser, ids)
		if detailErr != nil {
			return nil, detailErr
		}
		for _, u := range users {
			isFriend := data.Contains(friends, u.ID) // no DB query
			event := &data.UserEvent{
				Profile:   u,
				EventDate: isoDate,
				EventType: eventType,
				EventUser: forUser.Username,
				IsFriend:  isFriend,
			}
			list = append(list, event)
		}
	}
	return
}
