# Facebook Scraper

Scrapes public Facebook data including pages, posts, events, videos, reviews, and ads library.

## Installation

```bash
go get github.com/embedtools/facebook-scraper
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/embedtools/facebook-scraper/scraper"
	"github.com/embedtools/facebook-scraper/types"
)

func main() {
	client, err := scraper.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	_ = client
	_ = ctx
	fmt.Println("Facebook scraper ready")
}
```

## Capabilities

| Capability | Method | Type | Timeout |
|---|---|---|---|
| facebook.page.get | GetPage | sync | 30s |
| facebook.page-posts.list | ListPagePosts | streaming | 120s |
| facebook.post-comments.list | ListPostComments | streaming | 120s |
| facebook.event.get | GetEvent | sync | 30s |
| facebook.page-reviews.list | ListPageReviews | streaming | 120s |
| facebook.ads-library.search | SearchAdsLibrary | streaming | 120s |
| facebook.video.get | GetVideo | sync | 30s |
| facebook.reel.get | GetReel | sync | 30s |
| facebook.page-videos.list | ListPageVideos | streaming | 120s |


## License

See repository root for license information.
