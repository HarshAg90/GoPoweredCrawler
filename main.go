package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"

	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// import (
// 	"flag"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"
// 	"github.com/PuerkitoBio/goquery"
// 	"github.com/joho/godotenv"
// 	"github.com/robfig/cron/v3"
// 	"github.com/urfave/cli/v2"
// )

var GlobalIndex int = 0

func main() {
	baseUrl := "https://en.wikipedia.org/wiki/Dog"
	LinkCheck := link_check(baseUrl)

	if !LinkCheck {
		fmt.Println("Link is not valid")
		return
	}

	re := regexp.MustCompile(`https?://([^/]+)`)
	match := re.FindStringSubmatch(baseUrl)
	if len(match) > 1 {
		fmt.Printf("URL: %s → Domain: %s\n", baseUrl, match[1])
	} else {
		fmt.Printf("URL: %s → No match\n", baseUrl)
	}
	crawl(baseUrl)
}

// func crawl(PIndex string, link string) {
func crawl(link string) bool {
	res := link_check(link)

	if !res {
		fmt.Println("Link is not valid")
		return false
	}
	// index := 1 // current link index

	// request url
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("Failed to fetch %s: %v\n", link, err)
		return false
	}
	defer resp.Body.Close()

	// check for response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK HTTP status: %d for %s\n", resp.StatusCode, link)
		return false
	}

	// parse response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Failed to parse HTML for %s: %v\n", link, err)
		return false
	}

	// find all links
	var links []string
	linkSet := make(map[string]struct{})
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && link_check(href) {
			if _, found := linkSet[href]; !found {
				links = append(links, href)
				linkSet[href] = struct{}{}
			}
			// fmt.Printf("Found link: %s\n", href)
		}
	})
	// fmt.Printf("All found links: %v\n", links)

	// also check if link is Asset (image, video, audio, etc) and index them differently
	var assets []string
	doc.Find("img[src], video[src], audio[src]").Each(func(i int, s *goquery.Selection) {
		var src string
		if val, ok := s.Attr("src"); ok {
			src = val
			assets = append(assets, src)
		}
		// fmt.Printf("Found asset: %s\n", src)
	})

	// stored_sites_db[Pindex] := { "site": link, "body":siteContent,
	// 								{1: {link :link1, index: "{Pindex}"}, 2: link2, 3: link3}, {1: img1, 2: video, 3: image2} }
	htmlString, err := doc.Html()
	if err != nil {
		fmt.Printf("Failed to get HTML string for %s: %v\n", link, err)
		return false
	}
	storeSite(link, htmlString, links, assets)

	// this for websites only
	// links := []string{"https://example.com","https://example.com/about"}
	UniqueLinks := uniqueUnprocessedLinks(links)
	for i := 0; i < len(UniqueLinks); i++ {
		linkL := UniqueLinks[i]
		if linkL == link {
			return false
		}
		// argIndex := PIndex + "." + strconv.Itoa(index)
		if GlobalIndex < 20 {
			crawl(linkL) // async?
		} else {
			return false
		}
		// also need a async storage function so this doesnt hog memory
		// index++
	}
	return true
}

func storeSite(link string, bd any, links any, assets any) bool {
	siteData := map[string]interface{}{
		"link":   link,
		"body":   bd,
		"links":  links,
		"assets": assets,
		"metadata": map[string]interface{}{
			"timestamp": time.Now(),
			"status":    "crawled",
		},
	}
	// fmt.Printf("Site data: %v\n", siteData)

	// Set up MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Printf("Mongo client error: %v\n", err)
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Mongo connect error: %v\n", err)
		return false
	}
	defer client.Disconnect(ctx)

	// Check if collection with the name of the link already exists
	collections, err := client.Database("CrawledDb").ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		fmt.Printf("Mongo list collections error: %v\n", err)
		return false
	}
	for _, collName := range collections {
		if collName == link {
			fmt.Printf("skipping for'%s'.\n", link)
			return false
		}
	}

	collection := client.Database("CrawledDb").Collection(link)
	_, err = collection.InsertOne(ctx, siteData)
	if err != nil {
		fmt.Printf("Mongo insert error: %v\n", err)
		return false
	}
	fmt.Println("Site data stored in MongoDB")
	GlobalIndex++
	fmt.Printf("index: %v\n", GlobalIndex)
	return true
}
func link_check(link string) bool {
	re := regexp.MustCompile(`^(https?://)?([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}(/.*)?$`)
	return re.MatchString(link)
}

// Not using because of long time to check
func link_check1(link string) bool {
	re := regexp.MustCompile(`^https?://`)
	if !re.MatchString(link) {
		return false
	}

	client := http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Head(link)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Check for broken link (4xx or 5xx)
	if resp.StatusCode >= 400 {
		return false
	}

	// Optionally, check if it's a redirect
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		// It's a redirect, but still reachable
		return true
	}

	return true
}

func uniqueUnprocessedLinks(input []string) []string {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Printf("Mongo client error: %v\n", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Mongo connect error: %v\n", err)
		return nil
	}
	defer client.Disconnect(ctx)

	// Get all collection names (already processed links)
	collections, err := client.Database("CrawledDb").ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		fmt.Printf("Mongo list collections error: %v\n", err)
		return nil
	}
	processed := make(map[string]struct{})
	for _, collName := range collections {
		processed[collName] = struct{}{}
	}

	// Filter input for unique and unprocessed links
	seen := make(map[string]struct{})
	var result []string
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			if _, done := processed[v]; !done {
				result = append(result, v)
			}
		}
	}
	return result
}
