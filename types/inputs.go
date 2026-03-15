package types

// GetPageInput holds the data fields.
type GetPageInput struct {
	URL string `json:"url"`
}
// ListPagePostsInput holds the data fields.
type ListPagePostsInput struct {
	URL string `json:"url"`
	Limit int `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}
// ListPostCommentsInput holds the data fields.
type ListPostCommentsInput struct {
	URL string `json:"url"`
	Limit int `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}
// SearchAdsLibraryInput holds the data fields.
type SearchAdsLibraryInput struct {
	Query string `json:"query"`
	Limit int `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}
// GetVideoInput holds the data fields.
type GetVideoInput struct {
	URL string `json:"url"`
}
// GetReelInput holds the data fields.
type GetReelInput struct {
	URL string `json:"url"`
}
// ListPageVideosInput holds the data fields.
type ListPageVideosInput struct {
	URL string `json:"url"`
	Limit int `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

// GetPostInput holds the data fields.
type GetPostInput struct {
	URL string `json:"url"`
}

// GetPhotoInput holds the data fields.
type GetPhotoInput struct {
	URL string `json:"url"`
}

// ListPageReelsInput holds the data fields.
type ListPageReelsInput struct {
	URL   string `json:"url"`
	Limit int    `json:"limit,omitempty"`
}
