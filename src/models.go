package main

// RequestBody is the expected incoming JSON request
type RequestBody struct {
	Threads int      `json:"threads"`
	Urls    []string `json:"urls"`
}

// JobURLImageMap is a map with the jobID as the key and the value is another map with the URL and image links
type JobURLImageMap map[string]map[string][]string

// JobProgressMap is a map with the jobID as the key and number of processes as the value
type JobProgressMap map[string]int
