package app

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/pkg/errors"
)

func (a *App) reportHandler(c *gin.Context) {
	forUser, err := a.getUser(c)
	if err != nil {
		a.logger.Printf("error getting user from context: %v", err)
		a.logOutHandler(c)
		return
	}

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil || profile.Username == "" {
		a.viewErrorHandler(c, http.StatusBadRequest, err, "Error getting user profile")
		return
	}

	data := gin.H{
		"user":    profile,
		"version": a.appVersion,
	}

	c.HTML(http.StatusOK, "report", data)
}

func (a *App) reportDataHandler(c *gin.Context) {
	ctx := c.Request.Context()
	forUser, err := a.getUser(c)
	if err != nil {
		a.errJSONAndAbort(c, err)
		return
	}

	startIDStr := c.Param("id")
	if startIDStr == "" {
		startIDStr = "0"
	}
	var pageErr error
	startID, err := strconv.ParseInt(startIDStr, 10, 32)
	if pageErr != nil {
		a.errJSONAndAbort(c, errors.Wrap(pageErr, "error parsing start ID number"))
		return
	}

	state, err := a.getState(forUser.Username, format.ToISODate(time.Now().UTC()))
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrap(err, "error getting current user state"))
		return
	}

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil || profile.Username == "" {
		a.errJSONAndAbort(c, errors.Wrapf(err, "error getting user profile for %s", forUser.Username))
		return
	}

	// sort IDs
	friendIDs := state.Friends
	sort.Slice(friendIDs, func(i, j int) bool {
		return friendIDs[i] < friendIDs[j]
	})

	noFollowIDs := make([]int64, 0)
	for _, id := range friendIDs {
		// skip the ones already paged
		if startID > 0 && id <= startID {
			continue
		}
		// get relationship
		rel, err := a.twClient.GetRelationship(ctx, forUser, profile.ID, id)
		if err != nil {
			a.errJSONAndAbort(c, errors.Wrap(err, "error getting user relationship"))
			return
		}
		// add if not follows
		if !rel.Target.Following {
			noFollowIDs = append(noFollowIDs, id)
		}
		// finish on page size
		if len(noFollowIDs) >= a.pageSize {
			break
		}
	}

	users, err := a.twClient.GetUserDetailsFromIDs(ctx, forUser, noFollowIDs)
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrap(err, "error getting user details"))
		return
	}

	var lastID int64 = 0
	if len(users) > 0 {
		lastID = users[len(users)-1].ID
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       profile,
		"state":      state,
		"version":    a.appVersion,
		"updated_on": forUser.UpdatedAt,
		"list":       users,
		"startID":    startID,
		"hasMore":    len(noFollowIDs) >= a.pageSize,
		"lastID":     lastID,
	})
}
