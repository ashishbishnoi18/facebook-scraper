package types

import "encoding/json"

type GetPageOutput = json.RawMessage
type ListPagePostsItem = json.RawMessage

// ListPagePostsSummary holds pagination metadata.
type ListPagePostsSummary struct {
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type ListPostCommentsItem = json.RawMessage

// ListPostCommentsSummary holds pagination metadata.
type ListPostCommentsSummary struct {
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type SearchAdsLibraryItem = json.RawMessage

// SearchAdsLibrarySummary holds pagination metadata.
type SearchAdsLibrarySummary struct {
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type GetVideoOutput = json.RawMessage
type GetReelOutput = json.RawMessage
type ListPageVideosItem = json.RawMessage

// ListPageVideosSummary holds pagination metadata.
type ListPageVideosSummary struct {
	TotalItems int    `json:"total_items,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type GetPostOutput = json.RawMessage
type GetPhotoOutput = json.RawMessage
type ListPageReelsItem = json.RawMessage

// ListPageReelsSummary holds pagination metadata.
type ListPageReelsSummary struct {
	TotalItems int `json:"total_items,omitempty"`
}
