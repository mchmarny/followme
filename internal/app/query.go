package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/asdine/storm/v3"
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

	var state data.DailyState
	stateKey := data.GetDailyStateKey(forUser.Username, time.Now().UTC())
	if err := a.db.One("Key", stateKey, &state); err != nil {
		a.errJSONAndAbort(c, errors.Wrapf(err, "error getting user state for %s", stateKey))
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
		dayState, err := a.getState(forUser.Username, date)
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

		a.logger.Printf("day[%d] +:%d -%d a:%f ra:%f f+:%d f-:%d",
			day,
			dayState.NewFollowerCount,
			dayState.NewUnfollowerCount,
			runSum,
			series.AvgFollowers[dayState.StateOn],
			dayState.NewFriendsCount,
			dayState.NewUnfriendedCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       profile,
		"state":      state,
		"version":    a.appVersion,
		"updated_on": forUser.UpdatedAt,
		"days":       days,
		"series":     series,
	})
}

func (a *App) getState(username string, date time.Time) (*data.DailyState, error) {
	key := data.GetDailyStateKey(username, date)
	ds := format.ToISODate(date)
	var s data.DailyState
	if err := a.db.One("Key", key, &s); err != nil {
		if err != storm.ErrNotFound {
			return nil, errors.Wrapf(err, "error getting state for %s", key)
		}
		s = data.DailyState{
			Key:      key,
			Username: username,
			StateOn:  ds,
		}
	}
	return &s, nil
}

// errJSONAndAbort throws JSON error and abort prevents pending handlers from being called
func (a *App) errJSONAndAbort(c *gin.Context, err error) {
	a.logger.Printf("error while processing JSON request: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Internal server error, see logs for details",
		"status":  "Error",
	})
	c.Abort()
}
