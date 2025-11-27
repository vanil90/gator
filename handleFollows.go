package main

import (
	"context"
	"fmt"
	"gator/internal/database"
	"time"

	"github.com/google/uuid"
)

func printFollow(follows []database.CreateFeedFollowRow) {
	for _, f := range follows {
		fmt.Printf("ID:\t\t%s\nFeed Name:\t%s\nUser Name:\t%s\n", f.ID, f.FeedName, f.UserName)
	}
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("follow: invalid arguments, expected %d but got %d", 1, len(cmd.Args))
	}
	url := cmd.Args[0]
	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("follow: failed to find feed: %w", err)
	}

	follow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("follow: failed to create follow: %w", err)
	}

	printFollow(follow)

	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("following: invalid arguments, expected %d but got %d", 0, len(cmd.Args))
	}

	ctx := context.Background()
	follows, err := s.db.GetFeedFollowsForUser(ctx, user.Name)
	if err != nil {
		return fmt.Errorf("following: failed to find follows for current user: %w", err)
	}

	fmt.Printf("%s follows:\n", user.Name)
	for _, f := range follows {
		fmt.Printf("Feed Name:\t%s\nFeed URL:\t%s\n", f.FeedName, f.FeedUrl)
		fmt.Println("---")
	}

	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("unfollow: invalid arguments, expected %d but got %d", 1, len(cmd.Args))
	}
	url := cmd.Args[0]
	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("unfollow: feed not found: %w", err)
	}

	err = s.db.DeleteFollow(ctx, database.DeleteFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("unfollow: failed to delete feed: %w", err)
	}

	fmt.Printf("'%s' unfollowed successfully!\n", feed.Name)
	return nil
}
