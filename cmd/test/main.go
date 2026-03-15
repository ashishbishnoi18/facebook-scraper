package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/embedtools/facebook-scraper/scraper"
	"github.com/embedtools/facebook-scraper/types"
)

func main() {
	client, err := scraper.New()
	if err != nil {
		log.Fatalf("init: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test 1: GetPage
	fmt.Println("=== Test 1: GetPage (Meta) ===")
	page, err := client.GetPage(ctx, &types.GetPageInput{URL: "https://www.facebook.com/Meta"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetPage", *page)
	}

	// Test 2: GetPage (different page)
	fmt.Println("=== Test 2: GetPage (NASA) ===")
	page2, err := client.GetPage(ctx, &types.GetPageInput{URL: "https://www.facebook.com/NASA"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetPage", *page2)
	}

	// Test 3: ListPagePosts
	fmt.Println("=== Test 3: ListPagePosts (Meta, limit 3) ===")
	var posts []json.RawMessage
	summary, err := client.ListPagePosts(ctx, &types.ListPagePostsInput{URL: "https://www.facebook.com/Meta", Limit: 3}, func(item *types.ListPagePostsItem) error {
		posts = append(posts, *item)
		fmt.Printf("  Post emitted (%d bytes)\n", len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Summary: %d items, cursor: %q\n\n", summary.TotalItems, summary.NextCursor)
	}

	// Test 4: GetPost (a known public post URL)
	fmt.Println("=== Test 4: GetPost ===")
	post, err := client.GetPost(ctx, &types.GetPostInput{URL: "https://www.facebook.com/Meta/posts/pfbid02FLWLqhqNjMXpSCsDBKBLqcgRMGrVijVVoeU6MAsVeF7oUnw6WMLV17hdXsNjXl5Rl"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetPost", *post)
	}

	// Test 5: GetVideo
	fmt.Println("=== Test 5: GetVideo ===")
	video, err := client.GetVideo(ctx, &types.GetVideoInput{URL: "https://www.facebook.com/Meta/videos/1234567890"})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		printJSON("GetVideo", *video)
	}

	fmt.Println("=== Done ===")
}

func printJSON(label string, data json.RawMessage) {
	var pretty map[string]interface{}
	if err := json.Unmarshal(data, &pretty); err != nil {
		fmt.Printf("%s (raw): %s\n\n", label, string(data))
		return
	}
	b, _ := json.MarshalIndent(pretty, "", "  ")
	// Truncate if too long
	s := string(b)
	if len(s) > 2000 {
		s = s[:2000] + "\n  ... (truncated)"
	}
	fmt.Printf("%s:\n%s\n\n", label, s)
}
