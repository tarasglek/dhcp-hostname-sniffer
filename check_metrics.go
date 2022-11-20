package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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

func metricsLoop(ch <-chan map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for record := range ch {
		enc, err := json.Marshal(record)
		if err != nil {
			log.Fatalf("Could not marshal JSON: %s", err)
		}
		log.Println(string(enc))
	}

}
