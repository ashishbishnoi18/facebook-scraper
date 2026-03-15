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

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Test: ListPostComments on a Meta post
	fmt.Println("=== ListPostComments ===")
	var comments []json.RawMessage
	cSummary, err := client.ListPostComments(ctx, &types.ListPostCommentsInput{
		URL:   "https://www.facebook.com/Meta/posts/1234567890",
		Limit: 5,
	}, func(item *types.ListPostCommentsItem) error {
		comments = append(comments, *item)
		fmt.Printf("  Comment emitted (%d bytes)\n", len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Comments summary: %d items\n\n", cSummary.TotalItems)
	}

	// Test: ListPageReels
	fmt.Println("=== ListPageReels (Meta, limit 3) ===")
	var reels []json.RawMessage
	rSummary, err := client.ListPageReels(ctx, &types.ListPageReelsInput{
		URL:   "https://www.facebook.com/Meta",
		Limit: 3,
	}, func(item *types.ListPageReelsItem) error {
		reels = append(reels, *item)
		fmt.Printf("  Reel emitted (%d bytes)\n", len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Reels summary: %d items\n\n", rSummary.TotalItems)
	}

	// Test: SearchAdsLibrary
	fmt.Println("=== SearchAdsLibrary (query: 'meta', limit 3) ===")
	var ads []json.RawMessage
	aSummary, err := client.SearchAdsLibrary(ctx, &types.SearchAdsLibraryInput{
		Query: "meta",
		Limit: 3,
	}, func(item *types.SearchAdsLibraryItem) error {
		ads = append(ads, *item)
		fmt.Printf("  Ad emitted (%d bytes)\n", len(*item))
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n\n", err)
	} else {
		fmt.Printf("Ads summary: %d items\n\n", aSummary.TotalItems)
	}

	// Test: GetPage with more pages
	pages := []string{
		"https://www.facebook.com/cocacola",
		"https://www.facebook.com/nike",
		"https://www.facebook.com/samsung",
	}
	for _, p := range pages {
		fmt.Printf("=== GetPage (%s) ===\n", p)
		page, err := client.GetPage(ctx, &types.GetPageInput{URL: p})
		if err != nil {
			fmt.Printf("ERROR: %v\n\n", err)
			continue
		}
		var data map[string]interface{}
		json.Unmarshal(*page, &data)
		name, _ := data["name"].(string)
		id, _ := data["id"].(string)
		cat, _ := data["category_name"].(string)
		fmt.Printf("  ID: %s | Name: %s | Category: %s\n\n", id, name, cat)
	}

	fmt.Println("=== Done ===")
}
