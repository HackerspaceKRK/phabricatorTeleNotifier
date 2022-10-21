package main

import (
	"time"

	"github.com/uber/gonduit/requests"
)

// FeedQueryRequest is the request struct for feed.query.
type FeedQueryRequest struct {
	After            string `json:"after,omitempty"`
	View             string `json:"view,omitempty"`
	requests.Request        // Includes __conduit__ field needed for authentication.
}

type FeedQueryResponseItem struct {
	Class            string `json:"class"`
	Epoch            int    `json:"epoch"`
	AuthorPHID       string `json:"authorPHID"`
	ChronologicalKey string `json:"chronologicalKey"`
	ObjectPHID       string `json:"objectPHID"`
	Text             string `json:"text"`
}

type FeedItem struct {
	URL              string
	Title            string
	Time             string
	Author           string
	Type             string
	TypeName         string
	ChronologicalKey string
	Text             string `json:"text"`
	TimeData         time.Time
}
