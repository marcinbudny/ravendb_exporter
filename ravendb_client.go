package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var (
	client = &http.Client{
		Timeout: timeout,
	}
)

type stats struct {
	memory map[string]interface{}
}

func getStats() (*stats, error) {
	data, err := get("/admin/debug/memory/stats")
	if err != nil {
		return nil, err
	}

	return &stats{memory: data}, nil
}

func get(path string) (map[string]interface{}, error) {
	response, err := client.Get(ravenBaseURL + path)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(buf, &result); err != nil {
		return nil, err
	}

	return result, nil
}
