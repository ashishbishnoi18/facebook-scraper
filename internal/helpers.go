package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const BaseURL = "https://www.facebook.com"

const GooglebotUA = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

// FetchPage fetches a Facebook page using Googlebot UA and returns the HTML body.
func FetchPage(ctx context.Context, client *http.Client, pageURL string) (string, int, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			return "", 0, ctx.Err()
		}
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		body, status, err := fetchPageOnce(ctx, client, pageURL)
		if err == nil {
			return body, status, nil
		}
		lastErr = err
	}
	return "", 0, lastErr
}

func fetchPageOnce(ctx context.Context, client *http.Client, pageURL string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", GooglebotUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(body), resp.StatusCode, nil
}

// ExtractSJSData extracts all ScheduledServerJS JSON payloads from script[data-sjs] tags.
func ExtractSJSData(html string) []map[string]interface{} {
	var results []map[string]interface{}
	search := html
	for {
		idx := strings.Index(search, "data-sjs")
		if idx == -1 {
			break
		}
		closeIdx := strings.Index(search[idx:], ">")
		if closeIdx == -1 {
			break
		}
		contentStart := idx + closeIdx + 1
		endIdx := strings.Index(search[contentStart:], "</script>")
		if endIdx == -1 {
			break
		}
		content := strings.TrimSpace(search[contentStart : contentStart+endIdx])
		if len(content) > 0 && content[0] == '{' {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(content), &data); err == nil {
				results = append(results, data)
			}
		}
		search = search[contentStart+endIdx:]
	}
	return results
}

// DeepFind recursively searches for a map containing all specified keys.
func DeepFind(v interface{}, keys []string, maxDepth int) map[string]interface{} {
	if maxDepth <= 0 {
		return nil
	}
	switch val := v.(type) {
	case map[string]interface{}:
		hasAll := true
		for _, k := range keys {
			if _, ok := val[k]; !ok {
				hasAll = false
				break
			}
		}
		if hasAll {
			return val
		}
		for _, child := range val {
			if result := DeepFind(child, keys, maxDepth-1); result != nil {
				return result
			}
		}
	case []interface{}:
		for _, child := range val {
			if result := DeepFind(child, keys, maxDepth-1); result != nil {
				return result
			}
		}
	}
	return nil
}

// DeepFindAll finds all maps containing all specified keys.
func DeepFindAll(v interface{}, keys []string, maxDepth int) []map[string]interface{} {
	var results []map[string]interface{}
	deepFindAllHelper(v, keys, maxDepth, &results)
	return results
}

func deepFindAllHelper(v interface{}, keys []string, maxDepth int, results *[]map[string]interface{}) {
	if maxDepth <= 0 {
		return
	}
	switch val := v.(type) {
	case map[string]interface{}:
		hasAll := true
		for _, k := range keys {
			if _, ok := val[k]; !ok {
				hasAll = false
				break
			}
		}
		if hasAll {
			*results = append(*results, val)
		}
		for _, child := range val {
			deepFindAllHelper(child, keys, maxDepth-1, results)
		}
	case []interface{}:
		for _, child := range val {
			deepFindAllHelper(child, keys, maxDepth-1, results)
		}
	}
}

// FindFeedEdges finds the timeline_list_feed_units edges from SJS scripts.
func FindFeedEdges(scripts []map[string]interface{}) ([]interface{}, bool) {
	for _, script := range scripts {
		feed := DeepFind(script, []string{"timeline_list_feed_units"}, 20)
		if feed != nil {
			tlfu := SubMap(feed, "timeline_list_feed_units")
			if tlfu != nil {
				edges := SubArray(tlfu, "edges")
				if edges != nil {
					return edges, true
				}
			}
		}
	}
	return nil, false
}

// FindCommentEdges finds comment edges from SJS scripts (post detail page).
func FindCommentEdges(scripts []map[string]interface{}) []interface{} {
	for _, script := range scripts {
		candidates := DeepFindAll(script, []string{"edges"}, 25)
		for _, c := range candidates {
			edges := SubArray(c, "edges")
			if len(edges) == 0 {
				continue
			}
			if edge0, ok := edges[0].(map[string]interface{}); ok {
				node := SubMap(edge0, "node")
				if node == nil {
					continue
				}
				if _, hasBody := node["body"]; hasBody {
					if _, hasAuthor := node["author"]; hasAuthor {
						if _, hasCT := node["created_time"]; hasCT {
							return edges
						}
					}
				}
			}
		}
	}
	return nil
}

// FindDelegatePage finds the richest delegate_page object from SJS scripts.
func FindDelegatePage(scripts []map[string]interface{}) map[string]interface{} {
	var best map[string]interface{}
	for _, script := range scripts {
		candidates := DeepFindAll(script, []string{"delegate_page"}, 20)
		for _, c := range candidates {
			dp := SubMap(c, "delegate_page")
			if dp == nil {
				continue
			}
			if _, hasName := dp["category_name"]; hasName {
				return dp
			}
			if best == nil || len(dp) > len(best) {
				best = dp
			}
		}
	}
	return best
}

// FindEventData finds event data from SJS scripts.
func FindEventData(scripts []map[string]interface{}) map[string]interface{} {
	for _, script := range scripts {
		ev := DeepFind(script, []string{"event_title", "start_timestamp"}, 25)
		if ev != nil {
			return ev
		}
	}
	return nil
}

// FindVideoData finds video data from SJS scripts.
func FindVideoData(scripts []map[string]interface{}) map[string]interface{} {
	for _, script := range scripts {
		v := DeepFind(script, []string{"permalink_url", "is_playable"}, 20)
		if v != nil {
			return v
		}
	}
	// Fallback: find by video fields
	for _, script := range scripts {
		v := DeepFind(script, []string{"playable_duration_in_ms", "title"}, 20)
		if v != nil {
			return v
		}
	}
	return nil
}

// ExtractOGTags extracts Open Graph meta tags from HTML.
func ExtractOGTags(html string) map[string]string {
	tags := make(map[string]string)
	re := regexp.MustCompile(`<meta\s+(?:property|name)="([^"]+)"\s+content="([^"]*?)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		tags[m[1]] = DecodeHTMLEntities(m[2])
	}
	return tags
}

// DecodeHTMLEntities decodes common HTML entities.
func DecodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&#x27;", "'")
	s = strings.ReplaceAll(s, "&#039;", "'")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#xb7;", "·")
	s = strings.ReplaceAll(s, "&#x2019;", "'")
	return s
}

// ParseLikesFromOG parses likes count from OG description like "106,813,204 likes · 263,421 talking about this."
func ParseLikesFromOG(desc string) int64 {
	re := regexp.MustCompile(`([\d,]+)\s+likes?`)
	m := re.FindStringSubmatch(desc)
	if len(m) < 2 {
		return 0
	}
	return ParseCommaNumber(m[1])
}

// ParseTalkingAboutFromOG parses "talking about this" count from OG description.
func ParseTalkingAboutFromOG(desc string) int64 {
	re := regexp.MustCompile(`([\d,]+)\s+talking about`)
	m := re.FindStringSubmatch(desc)
	if len(m) < 2 {
		return 0
	}
	return ParseCommaNumber(m[1])
}

// ParseCommaNumber parses a comma-separated number like "106,813,204" to int64.
func ParseCommaNumber(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// ParsePageIDFromAL extracts numeric page ID from al:android:url tag like "fb://profile/100080376596424".
func ParsePageIDFromAL(alURL string) string {
	re := regexp.MustCompile(`fb://profile/(\d+)`)
	m := re.FindStringSubmatch(alURL)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

var (
	rePageURL  = regexp.MustCompile(`facebook\.com/([^/?#]+)`)
	reVideoURL = regexp.MustCompile(`(?:facebook\.com/watch/?\?v=(\d+)|facebook\.com/[^/]+/videos/(\d+)|fb\.watch/\w+)`)
	reReelURL  = regexp.MustCompile(`facebook\.com/reel/(\d+)`)
	reEventURL = regexp.MustCompile(`facebook\.com/events/(\d+)`)
	rePostURL  = regexp.MustCompile(`facebook\.com/([^/]+)/posts/([^/?#]+)`)
)

// ParsePageSlug extracts the page slug/username from a Facebook URL.
func ParsePageSlug(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if !strings.Contains(u.Host, "facebook.com") {
		return "", fmt.Errorf("not a Facebook URL: %s", rawURL)
	}
	path := strings.TrimRight(u.Path, "/")
	parts := strings.Split(path, "/")
	for _, p := range parts {
		if p != "" && p != "pg" && p != "pages" {
			return p, nil
		}
	}
	return "", fmt.Errorf("page slug not found in URL: %s", rawURL)
}

// ParseVideoID extracts the video ID from a Facebook video URL.
func ParseVideoID(rawURL string) (string, error) {
	// /watch/?v=ID
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if v := u.Query().Get("v"); v != "" {
		return v, nil
	}
	// /username/videos/ID/ or /reel/ID/
	m := reVideoURL.FindStringSubmatch(rawURL)
	if len(m) >= 3 {
		if m[1] != "" {
			return m[1], nil
		}
		if m[2] != "" {
			return m[2], nil
		}
	}
	m = reReelURL.FindStringSubmatch(rawURL)
	if len(m) >= 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("video ID not found in URL: %s", rawURL)
}

// ParseEventID extracts the event ID from a Facebook event URL.
func ParseEventID(rawURL string) (string, error) {
	m := reEventURL.FindStringSubmatch(rawURL)
	if len(m) >= 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("event ID not found in URL: %s", rawURL)
}

// ParsePostID extracts page slug and post pfbid from a Facebook post URL.
func ParsePostID(rawURL string) (string, string, error) {
	m := rePostURL.FindStringSubmatch(rawURL)
	if len(m) >= 3 {
		return m[1], m[2], nil
	}
	// Try story.php format
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}
	storyFBID := u.Query().Get("story_fbid")
	id := u.Query().Get("id")
	if storyFBID != "" {
		return id, storyFBID, nil
	}
	return "", "", fmt.Errorf("post ID not found in URL: %s", rawURL)
}

// BuildPageURL builds a canonical page URL.
func BuildPageURL(slug string) string {
	return BaseURL + "/" + slug + "/"
}

// BuildPostURL builds a post URL.
func BuildPostURL(slug, postID string) string {
	return BaseURL + "/" + slug + "/posts/" + postID
}

// FormatTimestamp converts Unix timestamp to RFC3339.
func FormatTimestamp(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

// SubMap safely gets a nested map from a map.
func SubMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	sub, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return sub
}

// SubArray safely gets a nested array from a map.
func SubArray(m map[string]interface{}, key string) []interface{} {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

// Str safely gets a string from a map.
func Str(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

// Int64Val safely gets an int64 from a map.
func Int64Val(m map[string]interface{}, key string) int64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	}
	return 0
}

// Float64Val safely gets a float64 from a map.
func Float64Val(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

// DeepFindByKey finds the first value for a given key anywhere in the nested structure.
func DeepFindByKey(v interface{}, key string, maxDepth int) interface{} {
	if maxDepth <= 0 {
		return nil
	}
	switch val := v.(type) {
	case map[string]interface{}:
		if result, ok := val[key]; ok {
			return result
		}
		for _, child := range val {
			if result := DeepFindByKey(child, key, maxDepth-1); result != nil {
				return result
			}
		}
	case []interface{}:
		for _, child := range val {
			if result := DeepFindByKey(child, key, maxDepth-1); result != nil {
				return result
			}
		}
	}
	return nil
}
