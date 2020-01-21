package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/chilts/sid"
	"github.com/gin-gonic/gin"
)

var error404 = "JobID not found"

// HandleGETStatus is a route handler the status of a crawler job
func HandleGETStatus(c *gin.Context) {
	jobID := c.Param("jobID")

	if JobStorage[jobID] == nil {
		c.JSON(404, gin.H{
			"message": error404,
		})
		return
	}

	urlCount := 0
	for range JobStorage[jobID] {
		urlCount++
	}

	c.JSON(200, gin.H{
		"completed":  urlCount,
		"inprogress": JobProgressStorage[jobID],
	})
}

// HandleGETResult is a route handler that returns result of a crawler job
func HandleGETResult(c *gin.Context) {
	jobID := c.Param("jobID")

	if JobStorage[jobID] == nil {
		c.JSON(404, gin.H{
			"message": error404,
		})
		return
	}

	responses := make([]map[string][]string, 0)

	for url, images := range JobStorage[jobID] {
		response := make(map[string][]string)
		response[url] = images
		responses = append(responses, response)
	}

	data, _ := json.Marshal(responses)
	fmt.Println(string(data))

	c.JSON(200, responses)
}

// HandlePOST is a route handler that triggers a new crawler job
func HandlePOST(c *gin.Context) {
	rawRequestBody, _ := ioutil.ReadAll(c.Request.Body)
	var requestBody RequestBody

	decoder := json.NewDecoder(bytes.NewReader(rawRequestBody))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestBody)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	jobID := sid.IdBase32()

	threads := requestBody.Threads

	if threads <= 0 {
		threads = 1
	}

	JobProgressStorage[jobID] = 0

	go func() {
		for _, url := range requestBody.Urls {
			scrape(url, JobStorage, jobID, threads)
		}
	}()

	c.JSON(200, gin.H{
		"jobID":   jobID,
		"threads": threads,
		"urls":    requestBody.Urls,
	})
}
