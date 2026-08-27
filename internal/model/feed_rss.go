package model

import "time"

type FeedResponse struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Link        string     `json:"link"`
	Items       []FeedItem `json:"items"`
}

type FeedItem struct {
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	Content         string    `json:"content,omitempty"`
	Link            string    `json:"link"`
	Author          string    `json:"author,omitempty"`
	Published       string    `json:"published,omitempty"`
	PublishedParsed time.Time `json:"publishedParsed"`
}
