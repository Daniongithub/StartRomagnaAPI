package model

import "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"

type RTFeedType string

type RTFeed struct {
	Area string
	Type RTFeedType
	URL  string
	Feed *gtfs.FeedMessage
}
