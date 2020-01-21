package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var error404 = "JobID not found"

func HandleGETStatus(c *gin.Context) {
	jobID := c.Param("jobID")

	// make it generic function?
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

	//jobID := sid.IdBase32()
	jobID := strconv.Itoa(FakeJobID)
	FakeJobID += 1

	JobProgressStorage[jobID] = 0

	var wg sync.WaitGroup

	go func() {
		for _, url := range requestBody.Urls {
			scrape(url, JobStorage, jobID, requestBody.Threads, wg)
		}
	}()

	wg.Wait()

	c.JSON(200, gin.H{
		"jobID":   jobID,
		"threads": requestBody.Threads,
		"urls":    requestBody.Urls,
	})
}
