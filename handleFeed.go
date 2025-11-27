package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"gator/internal/database"
	"gator/internal/rss"
	"html"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func fetchFeed(ctx context.Context, feedUrl string) (*rss.RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result rss.RSSFeed
	dec := xml.NewDecoder(res.Body)
	err = dec.Decode(&result)
	if err != nil {
		return nil, err
	}

	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)

	return &result, nil
}

func handleAgg(s *State, cmd Command) error {
	rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(rss)
	return nil
}

func handleAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("addfeed: invalid arguments, expected %d but got %d", 2, len(cmd.Args))
	}

	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.Config.CurrentUsername)
	if err != nil {
		return fmt.Errorf("addfeed: %w", err)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("addfeed: %w", err)
	}

	s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	fmt.Println(feed)

	return nil
}

func handleListFeeds(s *State, cmd Command) error {
	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Name:\t%s\nURL:\t%s\nUser:\t%s\n", feed.FeedName, feed.FeedUrl, feed.UserName)
		fmt.Println("---")
	}
	return nil
}
