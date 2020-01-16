package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Go docker setup")

	r := gin.Default()
	r.GET("/status/:jobID", handleGETStatus)
	r.POST("/", handlePOST)
	r.Run()
}

func handleGETStatus(c *gin.Context) {

	jobID, err := strconv.Atoi(c.Param("jobID"))

	if err != nil {
		return
	}

	c.JSON(200, gin.H{
		"jobID":      jobID,
		"completed":  0,
		"inprogress": 0,
	})
}

func handlePOST(c *gin.Context) {

}
