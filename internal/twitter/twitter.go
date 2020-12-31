package twitter

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mchmarny/followme/pkg/pager"

	tw "github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/pkg/errors"
)

// NewTwitter creates a new instance of Twitter
func NewTwitter(key, secret string, logger *log.Logger) *Twitter {
	return &Twitter{
		oauthConfig: oauth1.NewConfig(key, secret),
		logger:      logger,
	}
}

// Twitter does Tiwtter things
type Twitter struct {
	oauthConfig *oauth1.Config
	logger      *log.Logger
}

func (t *Twitter) getClient(ctx context.Context, byUser *data.User) (client *tw.Client, err error) {
	token := oauth1.NewToken(byUser.AccessTokenKey, byUser.AccessTokenSecret)
	httpClient := t.oauthConfig.Client(oauth1.NoContext, token)
	return tw.NewClient(httpClient), nil
}

// GetUserDetails retreaves details about the user
func (t *Twitter) GetUserDetails(ctx context.Context, byUser *data.User) (user *data.Profile, err error) {
	// t.logger.Printf("User: %s", byUser.Username)
	users, err := t.getUsersByParams(ctx, byUser, &tw.UserLookupParams{
		ScreenName:      []string{byUser.Username},
		IncludeEntities: tw.Bool(true),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error quering Twitter for user: %s", byUser.Username)
	}
	if users == nil {
		return nil, fmt.Errorf("expected 1 user, found 0")
	}
	if len(users) != 1 {
		return nil, fmt.Errorf("expected 1 user, found ")
	}
	return users[0], nil
}

// GetUserDetailsFromIDs retreaves details about the user
func (t *Twitter) GetUserDetailsFromIDs(ctx context.Context, byUser *data.User, ids []int64) (users []*data.Profile, err error) {
	if byUser == nil {
		return nil, errors.Wrap(err, "user required")
	}
	if ids == nil {
		return nil, errors.Wrap(err, "ids required")
	}

	// t.logger.Printf("getting twitter profiles for %d ids", len(ids))

	p, err := pager.GetInt64ArrayPager(ids, 100, 0)
	if err != nil {
		return nil, errors.Wrap(err, "error creating pager")
	}
	for {
		list := p.Next()
		if list == nil {
			return
		}
		// t.logger.Printf("twitter profile request page: %d", len(list))
		u, err := t.getUsersByParams(ctx, byUser, &tw.UserLookupParams{
			UserID:          list,
			IncludeEntities: tw.Bool(true),
		})

		if err != nil {
			return nil, errors.Wrap(err, "error getting users")
		}

		// t.logger.Printf("twitter profile result page: %d", len(u))
		users = append(users, u...)
	}
}

func (t *Twitter) getUsersByParams(ctx context.Context, byUser *data.User, listParam *tw.UserLookupParams) (users []*data.Profile, err error) {
	client, err := t.getClient(ctx, byUser)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing client")
	}

	users = make([]*data.Profile, 0)
	items, resp, err := client.Users.Lookup(listParam)
	if err != nil {
		// TODO: find cleaner way of parsing error status code (17) from API error
		if resp.StatusCode == 404 && strings.Contains(err.Error(), "No user matches") {
			return users, nil
		}
		return nil, errors.Wrapf(err, "error paging followers (%s): %v", resp.Status, err)
	}

	for _, u := range items {
		usr := toSimpleUser(&u)
		users = append(users, usr)
	}

	return
}

func convertTwitterTime(v string) time.Time {
	t, err := time.Parse(time.RubyDate, v)
	if err != nil {
		t = time.Now()
	}
	return t.UTC()
}

func toSimpleUser(u *tw.User) *data.Profile {
	return &data.Profile{
		ID:            u.ID,
		Username:      format.NormalizeString(u.ScreenName),
		Name:          u.Name,
		Description:   u.Description,
		ProfileImage:  u.ProfileImageURLHttps,
		CreatedAt:     convertTwitterTime(u.CreatedAt),
		Following:     u.Following,
		Lang:          u.Lang,
		Location:      u.Location,
		Timezone:      u.Timezone,
		PostCount:     u.StatusesCount,
		FaveCount:     u.FavouritesCount,
		FriendCount:   u.FriendsCount,
		FollowerCount: u.FollowersCount,
		ListedCount:   u.ListedCount,
		UpdatedAt:     time.Now().UTC(),
	}
}

// GetFollowerIDs returns all follower IDs for authed user
func (t *Twitter) GetFollowerIDs(ctx context.Context, byUser *data.User) (ids []int64, err error) {
	client, err := t.getClient(ctx, byUser)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing client")
	}

	listParam := &tw.FollowerIDParams{
		ScreenName: byUser.Username,
		Count:      5000, // max per page
	}

	ids = make([]int64, 0)
	for {
		page, resp, err := client.Followers.IDs(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "error paging follower IDs (%s): %v", resp.Status, err)
		}

		// debug
		// logger.Printf("Page size:%d, Next:%d", len(page.IDs), page.NextCursor)

		ids = append(ids, page.IDs...)

		// has more IDs?
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return
}

// GetFriendIDs returns all IDs users following authed user
func (t *Twitter) GetFriendIDs(ctx context.Context, byUser *data.User) (ids []int64, err error) {
	client, err := t.getClient(ctx, byUser)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing client")
	}

	listParam := &tw.FriendIDParams{
		ScreenName: byUser.Username,
		Count:      5000, // max per page
	}

	ids = make([]int64, 0)
	for {
		page, resp, err := client.Friends.IDs(listParam)
		if err != nil {
			return nil, errors.Wrapf(err, "error paging following IDs (%s): %v", resp.Status, err)
		}

		// debug
		// logger.Printf("Page size:%d, Next:%d", len(page.IDs), page.NextCursor)

		ids = append(ids, page.IDs...)

		// has more IDs?
		if page.NextCursor < 1 {
			break
		}

		// reset cursor
		listParam.Cursor = page.NextCursor
	}

	return
}

// GetRelationship returns relationship between the source and the target
func (t *Twitter) GetRelationship(ctx context.Context, byUser *data.User, sourceID, targetID int64) (*tw.Relationship, error) {
	client, err := t.getClient(ctx, byUser)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing client")
	}

	params := &tw.FriendshipShowParams{
		SourceID: sourceID,
		TargetID: targetID,
	}

	rel, resp, err := client.Friendships.Show(params)
	if err != nil {
		return nil, errors.Wrapf(err, "error paging following IDs (%s): %v", resp.Status, err)
	}

	return rel, nil
}
