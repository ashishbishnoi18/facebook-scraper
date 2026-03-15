package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/embedtools/facebook-scraper/internal"
	"github.com/embedtools/facebook-scraper/types"
)

// Client is the Facebook scraper client.
type Client struct {
	http *http.Client
}

// New creates a new Facebook scraper client.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		http: http.DefaultClient,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func marshalRaw(data interface{}) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

// fetchAndParseSJS fetches a Facebook page and extracts SJS data.
func (c *Client) fetchAndParseSJS(ctx context.Context, pageURL string) (string, []map[string]interface{}, error) {
	html, status, err := internal.FetchPage(ctx, c.http, pageURL)
	if err != nil {
		return "", nil, fmt.Errorf("fetch failed: %w", err)
	}
	if status == 404 {
		return "", nil, ErrNotFound
	}
	if status == 429 {
		return "", nil, ErrRateLimited
	}
	if status >= 400 {
		return "", nil, fmt.Errorf("HTTP %d", status)
	}
	if strings.Contains(html, "login_form") && !strings.Contains(html, "data-sjs") {
		return "", nil, ErrBlocked
	}
	scripts := internal.ExtractSJSData(html)
	return html, scripts, nil
}

// GetPage implements capability facebook.page.get.
func (c *Client) GetPage(ctx context.Context, in *types.GetPageInput) (*types.GetPageOutput, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}
	slug, err := internal.ParsePageSlug(in.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	pageURL := internal.BuildPageURL(slug)
	html, scripts, err := c.fetchAndParseSJS(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	dp := internal.FindDelegatePage(scripts)
	if dp == nil {
		og := internal.ExtractOGTags(html)
		displayName := og["og:title"]
		id := internal.ParsePageIDFromAL(og["al:android:url"])
		if displayName == "" && id == "" {
			return nil, ErrNotFound
		}
		dp = map[string]interface{}{
			"display_name": displayName,
			"avatar_url":   og["og:image"],
			"url":          og["og:url"],
			"username":     slug,
			"id":           id,
		}
	}

	raw, err := marshalRaw(dp)
	if err != nil {
		return nil, err
	}
	out := types.GetPageOutput(raw)
	return &out, nil
}

// ListPagePosts implements capability facebook.page-posts.list.
func (c *Client) ListPagePosts(ctx context.Context, in *types.ListPagePostsInput, emit func(item *types.ListPagePostsItem) error) (*types.ListPagePostsSummary, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}
	slug, err := internal.ParsePageSlug(in.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	pageURL := internal.BuildPageURL(slug)
	_, scripts, err := c.fetchAndParseSJS(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	edges, found := internal.FindFeedEdges(scripts)
	if !found {
		return nil, ErrUpstreamChanged
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	emitted := 0
	for _, e := range edges {
		if emitted >= limit {
			break
		}
		edge, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		node := internal.SubMap(edge, "node")
		if node == nil {
			continue
		}

		postID := internal.Str(node, "post_id")
		if postID == "" {
			postID = internal.Str(node, "id")
		}
		if postID == "" {
			continue
		}

		raw, err := json.Marshal(node)
		if err != nil {
			continue
		}
		item := types.ListPagePostsItem(raw)
		if err := emit(&item); err != nil {
			return nil, err
		}
		emitted++
	}

	summary := &types.ListPagePostsSummary{
		TotalItems: emitted,
	}
	if len(edges) > 0 {
		lastEdge, ok := edges[len(edges)-1].(map[string]interface{})
		if ok {
			summary.NextCursor = internal.Str(lastEdge, "cursor")
		}
	}
	return summary, nil
}

// ListPostComments implements capability facebook.post-comments.list.
func (c *Client) ListPostComments(ctx context.Context, in *types.ListPostCommentsInput, emit func(item *types.ListPostCommentsItem) error) (*types.ListPostCommentsSummary, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}

	_, scripts, err := c.fetchAndParseSJS(ctx, in.URL)
	if err != nil {
		return nil, err
	}

	edges := internal.FindCommentEdges(scripts)
	if edges == nil {
		return &types.ListPostCommentsSummary{}, nil
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}

	emitted := 0
	for _, e := range edges {
		if emitted >= limit {
			break
		}
		edge, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		node := internal.SubMap(edge, "node")
		if node == nil {
			continue
		}

		raw, err := json.Marshal(node)
		if err != nil {
			continue
		}
		item := types.ListPostCommentsItem(raw)
		if err := emit(&item); err != nil {
			return nil, err
		}
		emitted++
	}

	return &types.ListPostCommentsSummary{
		TotalItems: emitted,
	}, nil
}


// SearchAdsLibrary implements capability facebook.ads-library.search.
func (c *Client) SearchAdsLibrary(ctx context.Context, in *types.SearchAdsLibraryInput, emit func(item *types.SearchAdsLibraryItem) error) (*types.SearchAdsLibrarySummary, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}
	if in.Query == "" {
		return nil, ErrInvalidURL
	}

	searchURL := fmt.Sprintf("%s/ads/library/?active_status=active&ad_type=all&country=ALL&q=%s", internal.BaseURL, url.QueryEscape(in.Query))
	html, scripts, err := c.fetchAndParseSJS(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	emitted := 0
	for _, script := range scripts {
		candidates := internal.DeepFindAll(script, []string{"ad_archive_id", "page_name"}, 25)
		for _, node := range candidates {
			if emitted >= limit {
				break
			}
			raw, err := json.Marshal(node)
			if err != nil {
				continue
			}
			item := types.SearchAdsLibraryItem(raw)
			if err := emit(&item); err != nil {
				return nil, err
			}
			emitted++
		}
		if emitted >= limit {
			break
		}
	}

	// Fallback: try finding ad cards by snapshot_url + page_id
	if emitted == 0 {
		for _, script := range scripts {
			candidates := internal.DeepFindAll(script, []string{"snapshot_url", "page_id"}, 25)
			for _, node := range candidates {
				if emitted >= limit {
					break
				}
				raw, err := json.Marshal(node)
				if err != nil {
					continue
				}
				item := types.SearchAdsLibraryItem(raw)
				if err := emit(&item); err != nil {
					return nil, err
				}
				emitted++
			}
			if emitted >= limit {
				break
			}
		}
	}

	// If still nothing, try OG-based fallback
	if emitted == 0 {
		og := internal.ExtractOGTags(html)
		if title := og["og:title"]; title != "" {
			node := map[string]interface{}{
				"query":       in.Query,
				"title":       title,
				"description": og["og:description"],
			}
			raw, _ := json.Marshal(node)
			item := types.SearchAdsLibraryItem(raw)
			if err := emit(&item); err != nil {
				return nil, err
			}
			emitted++
		}
	}

	return &types.SearchAdsLibrarySummary{
		TotalItems: emitted,
	}, nil
}

// GetVideo implements capability facebook.video.get.
func (c *Client) GetVideo(ctx context.Context, in *types.GetVideoInput) (*types.GetVideoOutput, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}

	html, scripts, err := c.fetchAndParseSJS(ctx, in.URL)
	if err != nil {
		return nil, err
	}

	var videoNode map[string]interface{}
	// Try rich Story node (video detail pages have post_id + actors + comet_sections)
	for _, script := range scripts {
		candidates := internal.DeepFindAll(script, []string{"post_id", "actors", "comet_sections"}, 20)
		for _, node := range candidates {
			videoNode = node
			break
		}
		if videoNode != nil {
			break
		}
	}
	// Fallback: try creation_time + post_id
	if videoNode == nil {
		for _, script := range scripts {
			candidates := internal.DeepFindAll(script, []string{"creation_time", "post_id"}, 20)
			for _, node := range candidates {
				videoNode = node
				break
			}
			if videoNode != nil {
				break
			}
		}
	}

	if videoNode == nil {
		og := internal.ExtractOGTags(html)
		title := og["og:title"]
		id, _ := internal.ParseVideoID(in.URL)
		if title == "" && id == "" {
			return nil, ErrNotFound
		}
		videoNode = map[string]interface{}{
			"title":         title,
			"description":   og["og:description"],
			"thumbnail_url": og["og:image"],
			"url":           og["og:url"],
			"id":            id,
		}
	}

	raw, err := marshalRaw(videoNode)
	if err != nil {
		return nil, err
	}
	out := types.GetVideoOutput(raw)
	return &out, nil
}

// GetReel implements capability facebook.reel.get.
func (c *Client) GetReel(ctx context.Context, in *types.GetReelInput) (*types.GetReelOutput, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}

	videoOut, err := c.GetVideo(ctx, &types.GetVideoInput{URL: in.URL})
	if err != nil {
		return nil, err
	}

	// videoOut is already json.RawMessage, just pass it through
	out := types.GetReelOutput(*videoOut)
	return &out, nil
}

// ListPageVideos implements capability facebook.page-videos.list.
func (c *Client) ListPageVideos(ctx context.Context, in *types.ListPageVideosInput, emit func(item *types.ListPageVideosItem) error) (*types.ListPageVideosSummary, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}
	slug, err := internal.ParsePageSlug(in.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	pageURL := internal.BuildPageURL(slug)
	_, scripts, err := c.fetchAndParseSJS(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	edges, found := internal.FindFeedEdges(scripts)
	if !found {
		return nil, ErrUpstreamChanged
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	emitted := 0
	for _, e := range edges {
		if emitted >= limit {
			break
		}
		edge, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		node := internal.SubMap(edge, "node")
		if node == nil {
			continue
		}

		// Check if this post has video attachments
		attachments := internal.SubArray(node, "attachments")
		hasVideo := false
		for _, att := range attachments {
			attMap, ok := att.(map[string]interface{})
			if !ok {
				continue
			}
			media := internal.SubMap(attMap, "media")
			if media != nil {
				typename := internal.Str(media, "__typename")
				if typename == "Video" {
					hasVideo = true
					break
				}
			}
			styles := internal.SubMap(attMap, "styles")
			if styles != nil {
				att2 := internal.SubMap(styles, "attachment")
				if att2 != nil {
					subs := internal.SubMap(att2, "all_subattachments")
					if subs != nil {
						nodes := internal.SubArray(subs, "nodes")
						for _, sn := range nodes {
							snMap, ok := sn.(map[string]interface{})
							if !ok {
								continue
							}
							m := internal.SubMap(snMap, "media")
							if m != nil && internal.Str(m, "__typename") == "Video" {
								hasVideo = true
								break
							}
						}
					}
				}
			}
		}
		if !hasVideo {
			continue
		}

		raw, err := json.Marshal(node)
		if err != nil {
			continue
		}
		item := types.ListPageVideosItem(raw)
		if err := emit(&item); err != nil {
			return nil, err
		}
		emitted++
	}

	return &types.ListPageVideosSummary{
		TotalItems: emitted,
	}, nil
}

// GetPost implements capability facebook.post.get.
func (c *Client) GetPost(ctx context.Context, in *types.GetPostInput) (*types.GetPostOutput, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}

	html, scripts, err := c.fetchAndParseSJS(ctx, in.URL)
	if err != nil {
		return nil, err
	}

	var postNode map[string]interface{}
	for _, script := range scripts {
		candidates := internal.DeepFindAll(script, []string{"post_id", "permalink_url", "actors"}, 20)
		for _, node := range candidates {
			postNode = node
			break
		}
		if postNode != nil {
			break
		}
	}

	if postNode == nil {
		og := internal.ExtractOGTags(html)
		text := og["og:description"]
		imageURL := og["og:image"]
		if text == "" && imageURL == "" {
			return nil, ErrNotFound
		}
		postNode = map[string]interface{}{
			"text":      text,
			"image_url": imageURL,
			"url":       og["og:url"],
		}
	}

	raw, err := marshalRaw(postNode)
	if err != nil {
		return nil, err
	}
	out := types.GetPostOutput(raw)
	return &out, nil
}

// GetPhoto implements capability facebook.photo.get.
func (c *Client) GetPhoto(ctx context.Context, in *types.GetPhotoInput) (*types.GetPhotoOutput, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}

	html, scripts, err := c.fetchAndParseSJS(ctx, in.URL)
	if err != nil {
		return nil, err
	}

	var media map[string]interface{}
	for _, script := range scripts {
		cm := internal.DeepFind(script, []string{"currMedia"}, 20)
		if cm == nil {
			continue
		}
		media = internal.SubMap(cm, "currMedia")
		if media != nil {
			break
		}
	}

	if media == nil {
		og := internal.ExtractOGTags(html)
		imageURL := og["og:image"]
		if imageURL == "" {
			return nil, ErrNotFound
		}
		media = map[string]interface{}{
			"image_url": imageURL,
			"url":       og["og:url"],
		}
	}

	raw, err := marshalRaw(media)
	if err != nil {
		return nil, err
	}
	out := types.GetPhotoOutput(raw)
	return &out, nil
}

// ListPageReels implements capability facebook.page-reels.list.
func (c *Client) ListPageReels(ctx context.Context, in *types.ListPageReelsInput, emit func(item *types.ListPageReelsItem) error) (*types.ListPageReelsSummary, error) {
	if ctx.Err() != nil {
		return nil, ErrContextCanceled
	}
	slug, err := internal.ParsePageSlug(in.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	reelsURL := internal.BaseURL + "/" + slug + "/reels/"
	_, scripts, err := c.fetchAndParseSJS(ctx, reelsURL)
	if err != nil {
		return nil, err
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	emitted := 0
	for _, script := range scripts {
		candidates := internal.DeepFindAll(script, []string{"shareable_url", "playback_video"}, 25)
		for _, reel := range candidates {
			if emitted >= limit {
				break
			}

			raw, err := json.Marshal(reel)
			if err != nil {
				continue
			}

			// Check for ID from URL
			shareURL := internal.Str(reel, "shareable_url")
			id := ""
			if shareURL != "" {
				parts := strings.Split(strings.TrimRight(shareURL, "/"), "/")
				if len(parts) > 0 {
					id = parts[len(parts)-1]
				}
			}
			if video := internal.SubMap(reel, "video"); video != nil {
				if vid := internal.Str(video, "id"); vid != "" {
					id = vid
				}
			}
			if id == "" {
				continue
			}

			item := types.ListPageReelsItem(raw)
			if err := emit(&item); err != nil {
				return nil, err
			}
			emitted++
		}
	}

	return &types.ListPageReelsSummary{
		TotalItems: emitted,
	}, nil
}
