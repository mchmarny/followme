package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/date"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/pkg/errors"
)

type dashboardSeries struct {
	AllFollowers  map[string]int     `json:"all_followers"`
	NewFollowers  map[string]int     `json:"new_followers"`
	LostFollowers map[string]int     `json:"lost_followers"`
	AvgFollowers  map[string]float32 `json:"avg_followers"`
	AvgTotal      map[string]float32 `json:"avg_total"`
	AllFriends    map[string]int     `json:"all_friends"`
	NewFriends    map[string]int     `json:"new_friends"`
	LostFriends   map[string]int     `json:"lost_friends"`
}

func (a *App) dashboardHandler(c *gin.Context) {
	profile, err := a.getUserProfile(c)
	if err != nil {
		a.logger.Printf("error getting profile: %v", err)
		a.logOutHandler(c)
		return
	}

	c.HTML(http.StatusOK, "dash", gin.H{
		"user":    profile,
		"version": a.appVersion,
		"refresh": c.Query("refresh"),
	})
}

func (a *App) dashboardQueryHandler(c *gin.Context) {
	forUser, err := a.getUser(c)
	if err != nil {
		a.errJSONAndAbort(c, err)
		return
	}

	daysStr := c.Query("days")
	if daysStr == "" {
		daysStr = "3"
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrapf(err, "error parsing days from '%s'", daysStr))
		return
	}

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil {
		a.errJSONAndAbort(c, errors.Wrapf(err, "error getting user profile for %s", forUser.Username))
		return
	}

	state, err := a.getState(forUser.Username, format.ToISODate(time.Now().UTC()))
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrap(err, "error getting current user state"))
		return
	}

	series := &dashboardSeries{
		AllFollowers:  map[string]int{},
		NewFollowers:  map[string]int{},
		LostFollowers: map[string]int{},
		AvgFollowers:  map[string]float32{},
		AvgTotal:      map[string]float32{},
		AllFriends:    map[string]int{},
		NewFriends:    map[string]int{},
		LostFriends:   map[string]int{},
	}

	var runSum float32 = 0
	var totalAvg float32 = 0

	for i, date := range date.GetDateRange(time.Now().UTC().AddDate(0, 0, -days)) {
		day := i + 1
		dayState, err := a.getState(forUser.Username, format.ToISODate(date))
		if err != nil {
			a.errJSONAndAbort(c, errors.Wrapf(err, "error getting user state for %v", date))
			return
		}

		// total
		series.AllFollowers[dayState.StateOn] = dayState.FollowerCount
		// followers (+/-)
		series.NewFollowers[dayState.StateOn] = dayState.NewFollowerCount
		series.LostFollowers[dayState.StateOn] = -dayState.NewUnfollowerCount
		// friend (+/-)
		series.NewFriends[dayState.StateOn] = dayState.NewFriendsCount
		series.LostFriends[dayState.StateOn] = -dayState.NewUnfriendedCount
		// avg
		runSum += float32(dayState.NewFollowerCount - dayState.NewUnfollowerCount)
		series.AvgFollowers[dayState.StateOn] = runSum / float32(day)

		// total avg
		totalAvg += float32(dayState.FollowerCount)
		series.AvgTotal[dayState.StateOn] = totalAvg / float32(day)

		// a.logger.Printf("day[%d] +:%d -%d a:%f ra:%f f+:%d f-:%d",
		// 	day,
		// 	dayState.NewFollowerCount,
		// 	dayState.NewUnfollowerCount,
		// 	runSum,
		// 	series.AvgFollowers[dayState.StateOn],
		// 	dayState.NewFriendsCount,
		// 	dayState.NewUnfriendedCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       profile,
		"state":      state,
		"version":    a.appVersion,
		"updated_on": state.UpdatedOn.Format(time.RFC1123),
		"days":       days,
		"series":     series,
	})
}
