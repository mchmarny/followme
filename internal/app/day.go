package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/list"
	"github.com/pkg/errors"
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
	if dayState.NewFollowerCount > a.maxEventLimit {
		newFollowerIDs = newFollowerIDs[len(newFollowerIDs)-a.maxEventLimit:]
		newFollowersLimit = true
	}

	followers, err := a.toUserEvent(ctx, forUser, profile.ID, dayState.Friends, newFollowerIDs, isoDate, data.FollowedEventType, false)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting new follower events")
		return
	}
	a.logger.Printf("Followers:%d, expected:%d", len(followers), dayState.FollowerCount)

	newUnfollowerIDs := dayState.NewUnfollowers
	var newUnfollowersLimit bool
	if dayState.NewUnfollowerCount > a.maxEventLimit {
		newUnfollowerIDs = newUnfollowerIDs[len(newUnfollowerIDs)-a.maxEventLimit:]
		newUnfollowersLimit = true
	}

	unfollowers, err := a.toUserEvent(ctx, forUser, profile.ID, dayState.Friends, newUnfollowerIDs, isoDate, data.UnfollowedEventType, false)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting unfollower events")
		return
	}
	a.logger.Printf("Unfollowers:%d, expected:%d", len(unfollowers), dayState.NewUnfollowerCount)

	newFriendedIDs := dayState.NewFriends
	var newFriendLimited bool
	if dayState.NewFriendsCount > a.maxEventLimit {
		newFriendedIDs = newFriendedIDs[len(newFriendedIDs)-a.maxEventLimit:]
		newFriendLimited = true
	}

	friended, err := a.toUserEvent(ctx, forUser, profile.ID, dayState.Friends, newFriendedIDs, isoDate, data.FriendedEventType, true)
	if err != nil {
		a.viewErrorHandler(c, http.StatusInternalServerError, err, "Error getting friended events")
		return
	}
	a.logger.Printf("Friends:%d, expected:%d", len(friended), dayState.NewFriendsCount)

	newUnfriendedIDs := dayState.NewUnfriended
	var newUnfriendLimited bool
	if dayState.NewUnfriendedCount > a.maxEventLimit {
		newUnfriendedIDs = newUnfriendedIDs[len(newUnfriendedIDs)-a.maxEventLimit:]
		newUnfriendLimited = true
	}

	unfriended, err := a.toUserEvent(ctx, forUser, profile.ID, dayState.Friends, newUnfriendedIDs, isoDate, data.UnfriendedEventType, true)
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
		"eventLimit": a.maxEventLimit,

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

func (a *App) toUserEvent(ctx context.Context, forUser *data.User, userID int64, friends, ids []int64, isoDate, eventType string, loadRel bool) (events []*data.UserEvent, err error) {
	a.logger.Printf("%s events for user %s - %d", eventType, forUser.Username, len(ids))
	if len(ids) > 0 {
		users, err := a.twClient.GetUserDetailsFromIDs(ctx, forUser, ids)
		if err != nil {
			return nil, errors.Wrap(err, "error getting user details")
		}
		a.logger.Printf("found: %d", len(users))
		for _, u := range users {
			event := &data.UserEvent{
				Profile:   u,
				EventDate: isoDate,
				EventType: eventType,
				EventUser: forUser.Username,
				IsFriend:  list.Contains(friends, u.ID),
			}

			if loadRel {
				// a.logger.Printf("user: %s", u.Username)
				rel, err := a.twClient.GetRelationship(ctx, forUser, userID, u.ID)
				if err != nil {
					return nil, errors.Wrap(err, "error getting user relationship")
				}
				a.logger.Printf("source:%d target:%d - %+v", userID, u.ID, rel)
				event.IsFollowing = rel.Target.Following
			}

			events = append(events, event)
		}
	}
	return
}
