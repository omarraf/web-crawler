package main

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omarraf/web-scraper/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"` // TODO: Add this field to the User struct
}

func databaseUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Name:      dbUser.Name,
		APIKey:    dbUser.ApiKey, // TODO: Add this field
	}
}

type Feed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

func databaseFeedToFeed(dbFeed database.Feed) Feed {
	return Feed{
		ID:        dbFeed.ID,
		CreatedAt: dbFeed.CreatedAt,
		UpdatedAt: dbFeed.UpdatedAt,
		Name:      dbFeed.Name,
		URL:       dbFeed.Url,
		UserID:    dbFeed.UserID,
	}
}

type FeedFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	FeedID    uuid.UUID `json:"feed_id"`
}

func databaseFeedFollowToFeedFollow(dbFeedFollow database.FeedFollow) FeedFollow {
	return FeedFollow{
		ID:        dbFeedFollow.ID,
		CreatedAt: dbFeedFollow.CreatedAt,
		UpdatedAt: dbFeedFollow.UpdatedAt,
		UserID:    dbFeedFollow.UserID,
		FeedID:    dbFeedFollow.FeedID,
	}
}

type Post struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Title       string    `json:"title"`
	Description *string   `json:"description"` // pointer because it's nullable
	PublishedAt time.Time `json:"published_at"`
	URL         string    `json:"url"`
	FeedID      uuid.UUID `json:"feed_id"`
}

func databasePostToPost(dbPost database.Post) Post {
	var description *string
	if dbPost.Description.Valid {
		description = &dbPost.Description.String
	}
	return Post{
		ID:          dbPost.ID,
		CreatedAt:   dbPost.CreatedAt,
		UpdatedAt:   dbPost.UpdatedAt,
		Title:       dbPost.Title,
		Description: description,
		PublishedAt: dbPost.PublishedAt,
		URL:         dbPost.Url,
		FeedID:      dbPost.FeedID,
	}
}

type CrawlJob struct {
	ID              uuid.UUID  `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UserID          uuid.UUID  `json:"user_id"`
	SeedURL         string     `json:"seed_url"`
	Status          string     `json:"status"`
	MaxDepth        int32      `json:"max_depth"`
	MaxPages        int32      `json:"max_pages"`
	PagesCrawled    int32      `json:"pages_crawled"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ErrorMsg        *string    `json:"error_msg,omitempty"`
	DiscoveredFeeds []string   `json:"discovered_feeds,omitempty"`
}

func databaseCrawlJobToCrawlJob(j database.CrawlJob) CrawlJob {
	out := CrawlJob{
		ID:           j.ID,
		CreatedAt:    j.CreatedAt,
		UpdatedAt:    j.UpdatedAt,
		UserID:       j.UserID,
		SeedURL:      j.SeedUrl,
		Status:       j.Status,
		MaxDepth:     j.MaxDepth,
		MaxPages:     j.MaxPages,
		PagesCrawled: j.PagesCrawled,
	}
	if j.StartedAt.Valid {
		t := j.StartedAt.Time
		out.StartedAt = &t
	}
	if j.FinishedAt.Valid {
		t := j.FinishedAt.Time
		out.FinishedAt = &t
	}
	if j.ErrorMsg.Valid {
		out.ErrorMsg = &j.ErrorMsg.String
	}
	if j.DiscoveredFeeds != "" {
		for _, f := range strings.Split(j.DiscoveredFeeds, "\n") {
			if f = strings.TrimSpace(f); f != "" {
				out.DiscoveredFeeds = append(out.DiscoveredFeeds, f)
			}
		}
	}
	return out
}

