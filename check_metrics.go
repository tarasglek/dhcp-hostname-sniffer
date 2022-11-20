package main

import (
	"fmt"
	"net/http"
	"time"
)

func discoverPrometheusEndpoint(ip string) bool {
	client := http.Client{
		Timeout: 1000 * time.Millisecond,
	}
	resp, err := client.Get("http://" + ip + "/metrics")
	if err != nil {
		fmt.Println("Error reaching out to endpoint " + ip + ": " + err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		fmt.Println("Endpoint not a valid prometheus endpoint")
		return false
	}
	return true
}
