package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func discoverPrometheusEndpoint(ip string) bool {
	client := http.Client{
		Timeout: 1000 * time.Millisecond,
	}
	url := "http://" + ip + "/metrics"
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reaching out to endpoint "+ip+": "+err.Error())
		return false
	}
	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, url+" not a valid prometheus endpoint")
		return false
	}
	return true
}

func metricsLoop(ch <-chan map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for record := range ch {
		if client_ip, ok := record["client_ip"].(string); ok {
			hasMetrics := discoverPrometheusEndpoint(client_ip)
			record["has_metrics"] = hasMetrics
		}
		enc, err := json.Marshal(record)
		if err != nil {
			log.Fatalf("Could not marshal JSON: %s", err)
		}
		log.Println(string(enc))
	}

}
