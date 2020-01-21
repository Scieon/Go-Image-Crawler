package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

// JobStorage is the in memory storage of jobs, urls scraped, and images found
var JobStorage = make(JobURLImageMap)

// JobProgressStorage is the in memory storage to keep track of number of running processes
var JobProgressStorage = make(JobProgressMap)

// URLImageMapMutex locks the Job Storage for concurrent writes
var URLImageMapMutex = sync.RWMutex{}

// JobProgressMapMutex locks the JobProgress storage for concurrent writes
var JobProgressMapMutex = sync.RWMutex{}

func main() {
	r := gin.Default()

	r.GET("/status/:jobID", HandleGETStatus)
	r.GET("/result/:jobID", HandleGETResult)
	r.POST("/", HandlePOST)

	r.Run()
}

// Scrapes a URL recursively
func scrape(url string, urlImageMap map[string]map[string][]string, jobID string, threads int) {

	var wg sync.WaitGroup

	c := colly.NewCollector(
		colly.MaxDepth(2),
	)

	// Buffered channel to limit number of goroutines
	concurrentGoroutines := make(chan struct{}, threads)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			e.Request.Visit(e.Attr("href"))
			<-concurrentGoroutines
		}()
	})

	c.OnRequest(func(r *colly.Request) {
		currentURL := fmt.Sprintf(r.URL.String())
		images := pullImages(currentURL, jobID)

		URLImageMapMutex.Lock()
		if urlImageMap[jobID] == nil {
			urlImageMap[jobID] = map[string][]string{}
		}
		urlImageMap[jobID][currentURL] = images
		URLImageMapMutex.Unlock()
	})

	JobProgressStorage[jobID]++
	c.Visit(url)
	wg.Wait()
	JobProgressStorage[jobID]--
}

// Returns slice of image links for a given URL
func pullImages(link string, jobID string) []string {
	var images []string

	JobProgressMapMutex.Lock()
	JobProgressStorage[jobID]++
	JobProgressMapMutex.Unlock()

	// Make HTTP request
	response, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		JobProgressStorage[jobID]--
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// Find and print image URLs
	document.Find("img").Each(func(index int, element *goquery.Selection) {
		imgSrc, exists := element.Attr("src")
		if exists {
			images = append(images, imgSrc)
		}
	})

	return images
}
