package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chilts/sid"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

func scrape() {
	c := colly.NewCollector(
		colly.MaxDepth(2),
	)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		go func() {
			e.Request.Visit(e.Attr("href"))
		}()
	})

	c.OnRequest(func(r *colly.Request) {
		go func() {
			fmt.Println("Visiting", r.URL)
			time.Sleep(5 * time.Second)
			pullImages("http://" + r.URL.Host)
		}()

	})

	c.Visit("http://go-colly.org/")
}

func pullImages(link string) {
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
			fmt.Println(imgSrc)
		}
	})
}

func main() {
	fmt.Println("Go docker setup")

	r := gin.Default()
	r.GET("/status/:jobID", handleGETStatus)
	r.POST("/", handlePOST)
	r.Run()
}

func handleGETStatus(c *gin.Context) {

	_, err := strconv.Atoi(c.Param("jobID"))

	if err != nil {
		return
	}

	// TODO should return array of urls
	scrape()

	id1 := sid.Id()

	// TODO make map with jobID: jobID, urls: [url1, url2]

	c.JSON(200, gin.H{
		"jobID":      id1,
		"completed":  0,
		"inprogress": 0,
	})
}

func handlePOST(c *gin.Context) {

}
