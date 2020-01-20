package main

import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Body struct {
	Threads int      `json:"threads"`
	Urls    []string `json:"urls"`
}

type jobObject map[string]map[string][]string

//{
//	"aweg-ea1ea": {
//		"http://golang.org": {
//			"http://gopher.png"
//		}
//	}
//}

var urlImageMap = make(jobObject)
var jobProgressMap = make(map[string]int)

var urlImageMapMutex = sync.RWMutex{}
var jobProgressMapMutex = sync.RWMutex{}
var fakeJobID = 1

func scrape(url string, urlImageMap map[string]map[string][]string, jobID string, threads int, wg sync.WaitGroup) {

	c := colly.NewCollector(
		colly.MaxDepth(2),
	)

	concurrentGoroutines := make(chan struct{}, threads)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			concurrentGoroutines <- struct{}{}
			//time.Sleep(2 * time.Second)
			//fmt.Println(e.Request.URL.String())

			e.Request.Visit(e.Attr("href")) //todo handle error
			<-concurrentGoroutines

		}()

	})

	c.OnRequest(func(r *colly.Request) {
		currentURL := fmt.Sprintf(r.URL.String())
		images := pullImages(currentURL, jobID)

		urlImageMapMutex.Lock()
		if urlImageMap[jobID] == nil {
			urlImageMap[jobID] = map[string][]string{}
		}
		urlImageMap[jobID][currentURL] = images
		urlImageMapMutex.Unlock()
	})

	jobProgressMap[jobID] += 1
	c.Visit(url)
	wg.Wait()
	jobProgressMap[jobID] -= 1
}

func pullImages(link string, jobID string) []string {
	var images []string

	jobProgressMapMutex.Lock()
	jobProgressMap[jobID] += 1
	jobProgressMapMutex.Unlock()

	// Make HTTP request
	response, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		jobProgressMap[jobID] -= 1
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

	//fmt.Println(len(images))
	return images
}

func main() {
	r := gin.Default()

	r.GET("/status/:jobID", handleGETStatus)
	r.GET("/result/:jobID", handleGETResult)
	r.POST("/", handlePOST)

	r.Run()
}

func handleGETStatus(c *gin.Context) {
	jobID := c.Param("jobID")

	// make it generic function?
	if urlImageMap[jobID] == nil {
		c.JSON(404, gin.H{
			"message": "Job ID not found",
		})
		return
	}

	urlCount := 0
	for range urlImageMap[jobID] {
		urlCount += 1
	}

	c.JSON(200, gin.H{
		"completed":  urlCount,
		"inprogress": jobProgressMap[jobID],
	})
}

func handleGETResult(c *gin.Context) {
	jobID := c.Param("jobID")

	if urlImageMap[jobID] == nil {
		c.JSON(404, gin.H{
			"message": "Job ID not found",
		})
		return
	}

	responses := make([]map[string][]string, 0)

	for url, images := range urlImageMap[jobID] {
		response := make(map[string][]string)
		response[url] = images
		responses = append(responses, response)
	}

	data, _ := json.Marshal(responses)
	fmt.Println(string(data))

	c.JSON(200, responses)
}

func handlePOST(c *gin.Context) {

	rawRequestBody, _ := ioutil.ReadAll(c.Request.Body)
	var requestBody Body

	// todo disallow unknown fields
	decoder := json.NewDecoder(bytes.NewReader(rawRequestBody))
	err := decoder.Decode(&requestBody)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Something went wrong",
		})
	}

	//jobID := sid.IdBase32()
	jobID := strconv.Itoa(fakeJobID)
	fakeJobID += 1

	jobProgressMap[jobID] = 0

	var wg sync.WaitGroup

	go func() {
		for _, url := range requestBody.Urls {
			scrape(url, urlImageMap, jobID, requestBody.Threads, wg)
		}
	}()

	wg.Wait()

	c.JSON(200, gin.H{
		"jobID":   jobID,
		"threads": requestBody.Threads,
		"urls":    requestBody.Urls,
	})
}
