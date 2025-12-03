package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

// RSS feed XML structure
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func urlToFeed(url string) (RSSFeed, error) {
	var rssFeed RSSFeed

	// Make an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return rssFeed, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return rssFeed, err
	}

	// Read the response body
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return rssFeed, err
	}

	// Parse XML
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return rssFeed, err
	}

	return rssFeed, nil
}

func parseRSSDate(dateStr string) (time.Time, error) {
	// Common RSS date format
	layout := "Mon, 02 Jan 2006 15:04:05 -0700"
	return time.Parse(layout, dateStr)
}
