package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/embedtools/facebook-scraper/scraper"
	"github.com/embedtools/facebook-scraper/types"
)

func main() {
	proxyURL, _ := url.Parse("http://ashishbishnoi18:Qw312OvD7jpdK7kr@proxy.packetstream.io:31112")
	proxiedClient := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			TLSHandshakeTimeout: 15 * time.Second,
		},
		Timeout: 90 * time.Second,
	}

	client, err := scraper.New(scraper.WithHTTPClient(proxiedClient))
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. GetPage
	fmt.Println("=== 1. GetPage (Meta) ===")
	testGetPage(ctx, client, "https://www.facebook.com/Meta")

	fmt.Println("=== 2. GetPage (NASA) ===")
	testGetPage(ctx, client, "https://www.facebook.com/NASA")

	fmt.Println("=== 3. GetPage (Samsung) ===")
	testGetPage(ctx, client, "https://www.facebook.com/samsung")

	// 4. ListPagePosts
	fmt.Println("=== 4. ListPagePosts (Meta, limit 3) ===")
	var postCount int
	summary, err := client.ListPagePosts(ctx, &types.ListPagePostsInput{URL: "https://www.facebook.com/Meta", Limit: 3}, func(item *types.ListPagePostsItem) error {
		postCount++
		fmt.Printf("  Post %d: %d bytes\n", postCount, len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d items, cursor: %q\n\n", summary.TotalItems, truncate(summary.NextCursor, 50))
	}

	// 5. ListPostComments — use a real post from Meta page
	fmt.Println("=== 5. ListPostComments (Meta post, limit 5) ===")
	var commentCount int
	cSummary, err := client.ListPostComments(ctx, &types.ListPostCommentsInput{
		URL:   "https://www.facebook.com/Meta/posts/pfbid0YDxRBsDMnQkWgJhfbkVHRWZSsXEV5VyBLK4bo6AEX7QfrdcCFoQ6xr9isnCJkJUl",
		Limit: 5,
	}, func(item *types.ListPostCommentsItem) error {
		commentCount++
		fmt.Printf("  Comment %d: %d bytes\n", commentCount, len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d comments\n\n", cSummary.TotalItems)
	}

	// 6. SearchAdsLibrary
	fmt.Println("=== 6. SearchAdsLibrary (query: 'coca cola', limit 3) ===")
	var adCount int
	aSummary, err := client.SearchAdsLibrary(ctx, &types.SearchAdsLibraryInput{Query: "coca cola", Limit: 3}, func(item *types.SearchAdsLibraryItem) error {
		adCount++
		fmt.Printf("  Ad %d: %d bytes\n", adCount, len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d ads\n\n", aSummary.TotalItems)
	}

	// 7. GetVideo
	fmt.Println("=== 7. GetVideo (real Meta video) ===")
	video, err := client.GetVideo(ctx, &types.GetVideoInput{URL: "https://www.facebook.com/watch/?v=1169493498175498"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetVideo", *video)
	}

	// 8. GetReel
	fmt.Println("=== 8. GetReel ===")
	reel, err := client.GetReel(ctx, &types.GetReelInput{URL: "https://www.facebook.com/reel/1234567890"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetReel", *reel)
	}

	// 9. ListPageReels
	fmt.Println("=== 9. ListPageReels (Meta, limit 3) ===")
	var reelCount int
	rSummary, err := client.ListPageReels(ctx, &types.ListPageReelsInput{URL: "https://www.facebook.com/Meta", Limit: 3}, func(item *types.ListPageReelsItem) error {
		reelCount++
		fmt.Printf("  Reel %d: %d bytes\n", reelCount, len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d reels\n\n", rSummary.TotalItems)
	}

	// 10. GetPhoto
	fmt.Println("=== 10. GetPhoto ===")
	photo, err := client.GetPhoto(ctx, &types.GetPhotoInput{URL: "https://www.facebook.com/photo/?fbid=266478569374694"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetPhoto", *photo)
	}

	// 11. GetPost
	fmt.Println("=== 11. GetPost ===")
	post, err := client.GetPost(ctx, &types.GetPostInput{URL: "https://www.facebook.com/Meta/posts/pfbid0YDxRBsDMnQkWgJhfbkVHRWZSsXEV5VyBLK4bo6AEX7QfrdcCFoQ6xr9isnCJkJUl"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetPost", *post)
	}

	// 12. ListPageVideos
	fmt.Println("=== 12. ListPageVideos (Meta, limit 3) ===")
	var vidCount int
	vSummary, err := client.ListPageVideos(ctx, &types.ListPageVideosInput{URL: "https://www.facebook.com/Meta", Limit: 3}, func(item *types.ListPageVideosItem) error {
		vidCount++
		fmt.Printf("  Video %d: %d bytes\n", vidCount, len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d videos\n\n", vSummary.TotalItems)
	}

	fmt.Println("=== ALL TESTS DONE ===")
}

func testGetPage(ctx context.Context, client *scraper.Client, pageURL string) {
	page, err := client.GetPage(ctx, &types.GetPageInput{URL: pageURL})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
		return
	}
	var data map[string]interface{}
	json.Unmarshal(*page, &data)
	name, _ := data["name"].(string)
	id, _ := data["id"].(string)
	cat, _ := data["category_name"].(string)
	desc := ""
	if bd, ok := data["best_description"].(map[string]interface{}); ok {
		desc, _ = bd["text"].(string)
	}
	fmt.Printf("  ID: %s | Name: %s | Category: %s\n", id, name, cat)
	if desc != "" {
		fmt.Printf("  Description: %s\n", truncate(desc, 100))
	}
	fmt.Println()
}

func printJSON(label string, data json.RawMessage) {
	var pretty interface{}
	json.Unmarshal(data, &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")
	s := string(b)
	if len(s) > 1500 {
		s = s[:1500] + "\n  ... (truncated)"
	}
	fmt.Printf("%s:\n%s\n\n", label, s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
