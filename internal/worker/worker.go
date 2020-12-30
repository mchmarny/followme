package worker

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/mchmarny/followme/internal/data"
	"github.com/mchmarny/followme/internal/twitter"
	"github.com/mchmarny/followme/pkg/format"
	"github.com/mchmarny/followme/pkg/list"
	"github.com/pkg/errors"
)

// NewWorker creates a new instance of the worker
func NewWorker(key, secret, url, version string) (*Worker, error) {
	if key == "" || secret == "" || version == "" {
		return nil, errors.New("key, secret, and version required")
	}

	// log
	logger := log.New(os.Stdout, "worker: ", 0)

	// data
	db, err := data.GetDB()
	if err != nil {
		return nil, errors.Wrap(err, "error getting DB")
	}

	// twitter
	t := twitter.NewTwitter(key, secret, logger)

	return &Worker{
		db:         db,
		twClient:   t,
		logger:     logger,
		appVersion: version,
	}, nil
}

// Worker represents the app worker
type Worker struct {
	db         *storm.DB
	twClient   *twitter.Twitter
	logger     *log.Logger
	appVersion string
}

func (w *Worker) updateUser(ctx context.Context, forUser data.User) error {
	if forUser.Username == "" {
		return errors.New("user parameter required")
	}

	w.logger.Printf("Starting processing for: %s...", forUser.Username)

	// ============================================================================
	// Twitter Details
	// ============================================================================
	userProfile, err := w.twClient.GetUserDetails(ctx, &forUser)
	if err != nil {
		return errors.Wrapf(err, "error getting twitter %s deails", forUser.Username)
	}

	if err := w.db.Save(userProfile); err != nil {
		return errors.Wrapf(err, "error saving %s profile", forUser.Username)
	}

	// ============================================================================
	// IDs of all followers from Twitter (users who follow this user)
	// ============================================================================
	w.logger.Println("Processing followers...")
	followerIDs, err := w.twClient.GetFollowerIDs(ctx, &forUser)
	if err != nil {
		return errors.Wrap(err, "error getting follower IDs")
	}
	w.logger.Printf("Follower counts for %s (Profile:%d, IDs:%d)",
		userProfile.Username, userProfile.FollowerCount, len(followerIDs))

	// ============================================================================
	// IDs of all friends from Twitter (users who this user follows)
	// ============================================================================
	w.logger.Println("Processing friends...")
	friendIDs, err := w.twClient.GetFriendIDs(ctx, &forUser)
	if err != nil {
		return errors.Wrap(err, "error getting friend IDs")
	}
	w.logger.Printf("Friend counts for %s (Profile:%d, IDs:%d)",
		userProfile.Username, userProfile.FriendCount, len(friendIDs))

	// ============================================================================
	// Yesterday State
	// ============================================================================
	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)

	yesterdayState, err := w.getState(forUser.Username, yesterday)
	if err != nil {
		return errors.Wrap(err, "error getting yesterday's state")
	}

	// ============================================================================
	//  Today State
	// ============================================================================
	todayState, err := w.getState(forUser.Username, today)
	if err != nil {
		return errors.Wrap(err, "error getting today's state")
	}

	// ============================================================================
	// New Followers
	// ============================================================================
	newFollowerIDs := list.GetDiff(yesterdayState.Followers, followerIDs)
	w.logger.Printf("New Followers (y:%d, +:%d)", yesterdayState.FollowerCount, len(newFollowerIDs))

	// ============================================================================
	// New Unfollowers
	// ============================================================================
	newUnfollowerIDs := list.GetDiff(followerIDs, yesterdayState.Followers)
	w.logger.Printf("New Unfollowers (y:%d, -:%d)", yesterdayState.FollowerCount, len(newUnfollowerIDs))

	// ============================================================================
	// New Friends
	// ============================================================================
	newFriendsIDs := list.GetDiff(yesterdayState.Friends, friendIDs)
	w.logger.Printf("New Friends (y:%d, +:%d)", yesterdayState.FriendsCount, len(newFriendsIDs))

	// ============================================================================
	// New Unfriends
	// ============================================================================
	newUnfriendsIDs := list.GetDiff(friendIDs, yesterdayState.Friends)
	w.logger.Printf("New Unfriends (y:%d, -:%d)", yesterdayState.FriendsCount, len(newUnfriendsIDs))

	// ============================================================================
	// Update State
	// ============================================================================
	todayState.Followers = followerIDs
	todayState.FollowerCount = len(followerIDs)

	todayState.NewFollowers = newFollowerIDs
	todayState.NewFollowerCount = len(newFollowerIDs)
	todayState.NewUnfollowers = newUnfollowerIDs
	todayState.NewUnfollowerCount = len(newUnfollowerIDs)

	todayState.Friends = friendIDs
	todayState.FriendsCount = len(friendIDs)

	todayState.NewFriends = newFriendsIDs
	todayState.NewFriendsCount = len(newFriendsIDs)
	todayState.NewUnfriended = newUnfriendsIDs
	todayState.NewUnfriendedCount = len(newUnfriendsIDs)

	// ============================================================================
	// Save State
	// ============================================================================
	w.logger.Printf("Saving updated state for %s", forUser.Username)
	if err := w.db.Save(todayState); err != nil {
		return errors.Wrap(err, "error saving daily state")
	}

	return nil

}

func (w *Worker) getState(username string, date time.Time) (*data.DailyState, error) {
	key := data.GetDailyStateKey(username, date)
	ds := format.ToISODate(date)
	var s data.DailyState
	if err := w.db.One("Key", key, &s); err != nil {
		if err != storm.ErrNotFound {
			return nil, errors.Wrap(err, "error getting yesterday's state")
		}
		s = data.DailyState{
			Key:      key,
			Username: username,
			StateOn:  ds,
		}
	}
	w.logger.Printf("%s state (follower:%d, +%d, -%d, friend:%d, +%d, -%d)", ds,
		s.FollowerCount, s.NewFollowerCount, s.NewUnfollowerCount,
		s.FriendsCount, s.NewFriendsCount, s.NewUnfriendedCount)
	return &s, nil
}

// Run run the update
func (w *Worker) Run() error {
	ctx := context.Background()
	w.logger.Println("Starting worker run...")

	var users []data.User
	if err := w.db.All(&users); err != nil {
		return errors.Wrap(err, "error while getting users")
	}
	w.logger.Printf("Found %d users", len(users))

	subErrors := 0
	for _, u := range users {
		if err := w.updateUser(ctx, u); err != nil {
			w.logger.Printf("error while updating user: %s - %v", u.Username, err)
			subErrors++
		}
	}

	if subErrors > 0 {
		return errors.Errorf("%d worker errors, see logs for details", subErrors)
	}

	return nil
}
