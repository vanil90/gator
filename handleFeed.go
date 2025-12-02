package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"gator/internal/database"
	"gator/internal/rss"
	"html"
	"net/http"
	"strconv"
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

func handleAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("agg: expected duration")
	}

	interval, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("agg: invalid duration: %w", err)
	}
	if interval < time.Second*10 {
		return errors.New("agg: invalid duration, should be at least 10s")
	}

	ticker := time.NewTicker(interval)
	for {
		scrapeFeeds(s)
		<-ticker.C
	}
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("addfeed: invalid arguments, expected %d but got %d", 2, len(cmd.Args))
	}

	ctx := context.Background()

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

func handleListFeeds(s *state, cmd command) error {
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

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("scrapefeed: %w", err)
	}

	err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("scrapefeed: %w", err)
	}

	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("scrapefeed: %w", err)
	}

	fmt.Printf("Fetched feed '%s'\n---\n", feed.Name)
	for _, item := range rssFeed.Channel.Item {
		pubDate, err := parseDate(item.PubDate)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Title: %s\nURL: %s\n***\n", pubDate.UTC(), item.Link)
		}

		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: pubDate,
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			FeedID:      feed.ID,
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	s.db.CreatePost(ctx, database.CreatePostParams{})
	fmt.Println("---")
	return nil
}

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

func parseDate(date string) (time.Time, error) {
	pubDate, err := time.Parse(dateLayout, date)
	if err != nil {
		return time.Time{}, err
	}

	return pubDate, nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("browse: too many arguments")
	}
	limit := 2
	var err error
	if len(cmd.Args) == 1 {
		limit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("browse: invalid limit: %w", err)
		}
	}

	ctx := context.Background()

	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("browse: failed to fetch posts for user")
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\nURL: %s\n***\n", post.Title, post.Url)
	}

	return nil
}
