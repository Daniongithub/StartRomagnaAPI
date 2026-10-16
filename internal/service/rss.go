package service

import (
	"errors"
	"startromagnaapi/internal/model"
	"time"

	"github.com/mmcdole/gofeed"
)

var ErrFeed = errors.New("rss feed error")

func GetRSSFeed(url string) (model.FeedResponse, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)
	if err != nil {
		return model.FeedResponse{}, ErrFeed
	}

	response := model.FeedResponse{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Items:       make([]model.FeedItem, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		author := ""
		if len(item.Authors) > 0 && item.Authors[0] != nil {
			author = item.Authors[0].Name
		}

		var published time.Time
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}
		response.Items = append(response.Items, model.FeedItem{
			Title:           item.Title,
			Description:     item.Description,
			Content:         item.Content,
			Link:            item.Link,
			Author:          author,
			Published:       item.Published,
			PublishedParsed: published,
		})
	}

	return response, nil
}
