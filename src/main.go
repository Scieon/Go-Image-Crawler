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

var JobStorage = make(JobURLImageMap)
var JobProgressStorage = make(JobProgressMap)
var UrlImageMapMutex = sync.RWMutex{}
var JobProgressMapMutex = sync.RWMutex{}
var FakeJobID = 1


// Scrapes a URL recursively
func scrape(url string, urlImageMap map[string]map[string][]string, jobID string, threads int, wg sync.WaitGroup) {

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
			//time.Sleep(2 * time.Second)
			//fmt.Println(e.Request.URL.String())
			e.Request.Visit(e.Attr("href"))
			<-concurrentGoroutines
		}()
	})

	c.OnRequest(func(r *colly.Request) {
		currentURL := fmt.Sprintf(r.URL.String())
		images := pullImages(currentURL, jobID)

		UrlImageMapMutex.Lock()
		if urlImageMap[jobID] == nil {
			urlImageMap[jobID] = map[string][]string{}
		}
		urlImageMap[jobID][currentURL] = images
		UrlImageMapMutex.Unlock()
	})

	JobProgressStorage[jobID] += 1
	c.Visit(url)
	wg.Wait()
	JobProgressStorage[jobID] -= 1
}

// Returns slice of image links for a given URL
func pullImages(link string, jobID string) []string {
	var images []string

	JobProgressMapMutex.Lock()
	JobProgressStorage[jobID] += 1
	JobProgressMapMutex.Unlock()

	// Make HTTP request
	response, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		JobProgressStorage[jobID] -= 1
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

func main() {
	r := gin.Default()

	r.GET("/status/:jobID", HandleGETStatus)
	r.GET("/result/:jobID", HandleGETResult)
	r.POST("/", HandlePOST)

	r.Run()
}
