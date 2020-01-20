package main

type RequestBody struct {
	Threads int      `json:"threads"`
	Urls    []string `json:"urls"`
}

type JobURLImageMap map[string]map[string][]string
type JobProgressMap map[string] int
