package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/mchmarny/followme/pkg/pager"
	"github.com/pkg/errors"
)

func (a *App) dayQueryHandler(c *gin.Context) {
	ctx := c.Request.Context()
	forUser, err := a.getUser(c)
	if err != nil {
		a.errJSONAndAbort(c, err)
		return
	}

	pageNum := 0
	pageStr := c.Param("page")
	if pageStr == "" {
		pageStr = "0"
	}
	var pageErr error
	pageNum, pageErr = strconv.Atoi(pageStr)
	if pageErr != nil {
		a.errJSONAndAbort(c, errors.Wrap(pageErr, "error parsing page number"))
		return
	}
	if pageNum < 0 {
		pageNum = 0
	}

	isoDate := c.Param("day")
	if isoDate == "" {
		a.errJSONAndAbort(c, errors.New("date required"))
		return
	}

	listType := c.Param("list")
	if listType == "" {
		a.errJSONAndAbort(c, errors.New("list type required"))
		return
	}

	state, err := a.getState(forUser.Username, isoDate)
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrap(err, "error getting current user state"))
		return
	}

	var profile data.Profile
	if err := a.db.One("Username", forUser.Username, &profile); err != nil || profile.Username == "" {
		a.errJSONAndAbort(c, errors.Wrapf(err, "error getting user profile for %s", forUser.Username))
		return
	}

	var (
		ids        []int64
		eventType  string
		followVerb string = "You Follow"
	)

	switch listType {
	case data.FollowedEventType:
		ids = state.NewFollowers
		eventType = data.FollowedEventType

	case data.UnfollowedEventType:
		ids = state.NewUnfollowers
		eventType = data.UnfollowedEventType

	case data.FriendedEventType:
		ids = state.NewFriends
		eventType = data.FriendedEventType
		followVerb = "Follows You"

	case data.UnfriendedEventType:
		ids = state.NewUnfriended
		eventType = data.UnfriendedEventType
		followVerb = "Follows You"

	default:
		a.errJSONAndAbort(c, errors.Wrapf(err, "invalid list type: %s", listType))
		return
	}

	// a.logger.Printf("all IDs:%d, page size:%d, page num:%d", len(ids), a.pageSize, pageNum)

	idPager, err := pager.GetInt64ArrayPager(ids, a.pageSize, pageNum)
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrapf(err,
			"error creating pager for %d items, page size:%d, page num:%d",
			len(ids), a.pageSize, pageNum))
		return
	}

	users, err := a.twClient.GetUserDetailsFromIDs(ctx, forUser, idPager.Next())
	if err != nil {
		a.errJSONAndAbort(c, errors.Wrap(err, "error getting user details"))
		return
	}

	// a.logger.Printf("users:%d", len(users))

	var events []*data.UserEvent
	for _, u := range users {
		event := &data.UserEvent{
			Profile:   u,
			EventDate: isoDate,
			EventType: eventType,
			EventUser: forUser.Username,
		}

		rel, err := a.twClient.GetRelationship(ctx, forUser, profile.ID, u.ID)
		if err != nil {
			a.errJSONAndAbort(c, errors.Wrap(err, "error getting user relationship"))
			return
		}

		if eventType == data.FollowedEventType || eventType == data.UnfollowedEventType {
			event.HasRelationship = format.ToYesNo(rel.Source.Following)
		} else {
			event.HasRelationship = format.ToYesNo(rel.Target.Following)
		}

		events = append(events, event)
	} // end for users

	//a.logger.Printf("events:%d", len(events))

	c.JSON(http.StatusOK, gin.H{
		"user":       profile,
		"state":      state,
		"version":    a.appVersion,
		"updated_on": forUser.UpdatedAt,
		"days":       isoDate,
		"events":     events,
		"listTyep":   listType,
		"pageNum":    pageNum,
		"pagePrev":   idPager.GetPrevPage(),
		"pageNext":   idPager.GetNextPage(),
		"hasPrev":    idPager.HasPrev(),
		"hasNext":    idPager.HasNext(),
		"followVerb": followVerb,
	})
}
