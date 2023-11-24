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

type TasksQueryRequest struct {
	requests.Request
	Status string `json:"status,omitempty"`
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

type PhabricatorTask struct {
	AuthorPHID         string   `json:"authorPHID"`
	Auxiliary          []any    `json:"auxiliary"`
	CcPHIDs            []string `json:"ccPHIDs"`
	DateCreated        string   `json:"dateCreated"`
	DateModified       string   `json:"dateModified"`
	DependsOnTaskPHIDs []any    `json:"dependsOnTaskPHIDs"`
	Description        string   `json:"description"`
	ID                 string   `json:"id"`
	IsClosed           bool     `json:"isClosed"`
	ObjectName         string   `json:"objectName"`
	OwnerPHID          string   `json:"ownerPHID"`
	Phid               string   `json:"phid"`
	Priority           string   `json:"priority"`
	PriorityColor      string   `json:"priorityColor"`
	ProjectPHIDs       []string `json:"projectPHIDs"`
	Status             string   `json:"status"`
	StatusName         string   `json:"statusName"`
	Title              string   `json:"title"`
	URI                string   `json:"uri"`
}

type ExtendedPhabricatorTask struct {
	PhabricatorTask
	AuthorName          string
	RenderedDescription any
	ProjectNames        []string
	IsImportant         bool
}

type GetFileInfoRequest struct {
	ID string `json:"id"`
	requests.Request
}

type GetFileInfoResp struct {
	ID           string `json:"id"`
	Phid         string `json:"phid"`
	ObjectName   string `json:"objectName"`
	Name         string `json:"name"`
	MimeType     string `json:"mimeType"`
	ByteSize     string `json:"byteSize"`
	AuthorPHID   string `json:"authorPHID"`
	DateCreated  string `json:"dateCreated"`
	DateModified string `json:"dateModified"`
	URI          string `json:"uri"`
}

type FileDownloadRequest struct {
	requests.Request
	Phid string `json:"phid"`
}

type FileDownloadResp struct {
	Result    string `json:"result"`
	ErrorCode any    `json:"error_code"`
	ErrorInfo any    `json:"error_info"`
}
