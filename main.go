package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/chilts/sid"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

type Body struct {
	Threads int `json:"threads"`
	Urls []string `json:"urls"`
}

var urlImageMap = make(map[string][]string)

func scrape(url string, urlImageMap map[string][]string) {
	c := colly.NewCollector(
		colly.MaxDepth(2),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// go func() {
		e.Request.Visit(e.Attr("href"))
		// }()
	})

	c.OnRequest(func(r *colly.Request) {
		// go func() {
		currentURL := fmt.Sprintf(r.URL.String())
		// time.Sleep(5 * time.Second)
		images := pullImages(currentURL)
		urlImageMap[currentURL] = images
		// }()

	})

	c.Visit(url)
}

func pullImages(link string) []string {
	var images []string

	// Make HTTP request
	response, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

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

	fmt.Println(len(images))
	return images
}

func main() {
	fmt.Println("Go docker setup")

	r := gin.Default()
	r.GET("/result/:jobID", handleGETStatus)
	r.POST("/", handlePOST)
	r.Run()
}

func handleGETStatus(c *gin.Context) {

	_, err := strconv.Atoi(c.Param("jobID"))

	if err != nil {
		return
	}

	c.JSON(200, gin.H{
		"blob":      urlImageMap,
	})
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

	for _, url := range requestBody.Urls {
		scrape(url, urlImageMap)
	}

	jobID := sid.Id()

	c.JSON(200, gin.H{
		"jobID": jobID,
		// todo get rid of debug fields
		"threads": requestBody.Threads,
		"url": requestBody.Urls,
		"blob": urlImageMap,
	})
}
