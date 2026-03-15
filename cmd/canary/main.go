package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/embedtools/facebook-scraper/scraper"
	"github.com/embedtools/facebook-scraper/types"
)

func main() {
	if os.Getenv("MODULE_CANARY_NETWORK") != "1" {
		fmt.Println("Skipping canary tests: MODULE_CANARY_NETWORK is not set to 1")
		os.Exit(0)
	}

	fmt.Println("=== Facebook Canary Runner ===")
	fmt.Println("Network tests enabled: true")
	fmt.Println()

	var opts []scraper.Option
	if proxy := os.Getenv("MODULE_PROXY"); proxy != "" {
		proxyURL, err := url.Parse("http://" + proxy)
		if err == nil {
			opts = append(opts, scraper.WithHTTPClient(&http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
				Timeout:   60 * time.Second,
			}))
			fmt.Printf("Using proxy: %s\n\n", proxy)
		}
	}

	client, err := scraper.New(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var total, passed, failed int

	// facebook.page.get — Meta page
	{
		fmt.Printf("  Testing facebook.page.get (GetPage)...\n")
		start := time.Now()
		input := &types.GetPageInput{
			URL: "https://www.facebook.com/Meta",
		}
		result, err := client.GetPage(ctx, input)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) len=%d\n", dur, len(*result))
			passed++
		}
		total++
	}

	// facebook.page-posts.list — Meta page posts
	{
		fmt.Printf("  Testing facebook.page-posts.list (ListPagePosts)...\n")
		start := time.Now()
		input := &types.ListPagePostsInput{
			URL:   "https://www.facebook.com/Meta",
			Limit: 5,
		}
		itemCount := 0
		_, err := client.ListPagePosts(ctx, input, func(item *types.ListPagePostsItem) error {
			itemCount++
			return nil
		})
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) items=%d\n", dur, itemCount)
			passed++
		}
		total++
	}

	// facebook.post-comments.list — specific Meta post
	{
		fmt.Printf("  Testing facebook.post-comments.list (ListPostComments)...\n")
		start := time.Now()
		input := &types.ListPostCommentsInput{
			URL:   "https://www.facebook.com/Meta/posts/pfbid0EwsxFj5K4PGG5XS58NxrZwEYgrx4aWwnSYJdFVsphLZDJMPXYiokrkefu1CmKm4Vl",
			Limit: 5,
		}
		itemCount := 0
		_, err := client.ListPostComments(ctx, input, func(item *types.ListPostCommentsItem) error {
			itemCount++
			return nil
		})
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) items=%d\n", dur, itemCount)
			passed++
		}
		total++
	}

	// facebook.ads-library.search
	{
		fmt.Printf("  Testing facebook.ads-library.search (SearchAdsLibrary)...\n")
		start := time.Now()
		input := &types.SearchAdsLibraryInput{
			Query: "Meta AI",
			Limit: 5,
		}
		itemCount := 0
		_, err := client.SearchAdsLibrary(ctx, input, func(item *types.SearchAdsLibraryItem) error {
			itemCount++
			return nil
		})
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) items=%d\n", dur, itemCount)
			passed++
		}
		total++
	}

	// facebook.video.get — Meta video
	{
		fmt.Printf("  Testing facebook.video.get (GetVideo)...\n")
		start := time.Now()
		input := &types.GetVideoInput{
			URL: "https://www.facebook.com/Meta/videos/911507208162880/",
		}
		result, err := client.GetVideo(ctx, input)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) len=%d\n", dur, len(*result))
			passed++
		}
		total++
	}

	// facebook.reel.get
	{
		fmt.Printf("  Testing facebook.reel.get (GetReel)...\n")
		start := time.Now()
		input := &types.GetReelInput{
			URL: "https://www.facebook.com/reel/911507208162880",
		}
		result, err := client.GetReel(ctx, input)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) len=%d\n", dur, len(*result))
			passed++
		}
		total++
	}

	// facebook.page-videos.list
	{
		fmt.Printf("  Testing facebook.page-videos.list (ListPageVideos)...\n")
		start := time.Now()
		input := &types.ListPageVideosInput{
			URL:   "https://www.facebook.com/Meta",
			Limit: 5,
		}
		itemCount := 0
		_, err := client.ListPageVideos(ctx, input, func(item *types.ListPageVideosItem) error {
			itemCount++
			return nil
		})
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) items=%d\n", dur, itemCount)
			passed++
		}
		total++
	}

	// facebook.post.get — specific Meta post
	{
		fmt.Printf("  Testing facebook.post.get (GetPost)...\n")
		start := time.Now()
		input := &types.GetPostInput{
			URL: "https://www.facebook.com/Meta/posts/pfbid0EwsxFj5K4PGG5XS58NxrZwEYgrx4aWwnSYJdFVsphLZDJMPXYiokrkefu1CmKm4Vl",
		}
		result, err := client.GetPost(ctx, input)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) len=%d\n", dur, len(*result))
			passed++
		}
		total++
	}

	// facebook.photo.get — Meta cover photo
	{
		fmt.Printf("  Testing facebook.photo.get (GetPhoto)...\n")
		start := time.Now()
		input := &types.GetPhotoInput{
			URL: "https://www.facebook.com/photo/?fbid=266478569374694",
		}
		result, err := client.GetPhoto(ctx, input)
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) len=%d\n", dur, len(*result))
			passed++
		}
		total++
	}

	// facebook.page-reels.list — Meta reels
	{
		fmt.Printf("  Testing facebook.page-reels.list (ListPageReels)...\n")
		start := time.Now()
		input := &types.ListPageReelsInput{
			URL:   "https://www.facebook.com/Meta",
			Limit: 5,
		}
		itemCount := 0
		_, err := client.ListPageReels(ctx, input, func(item *types.ListPageReelsItem) error {
			itemCount++
			return nil
		})
		dur := time.Since(start)
		if err != nil {
			fmt.Printf("    FAILED (%v): %v\n", dur, err)
			failed++
		} else {
			fmt.Printf("    PASSED (%v) items=%d\n", dur, itemCount)
			passed++
		}
		total++
	}

	fmt.Println()
	fmt.Printf("=== Summary ===\n")
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", total, passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}
