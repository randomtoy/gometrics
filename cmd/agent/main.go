package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

type Agent struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddr     string
	pollCount      int64
}

func NewAgent(serverAddr string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		serverAddr:     serverAddr,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
}

func (a *Agent) collectMetrics() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.pollCount++
	return map[string]interface{}{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),

		"PollCount":   a.pollCount,
		"RandomValue": rand.Float64(),
	}
}

func (a *Agent) sendMetrics(metrics map[string]interface{}) {
	for name, value := range metrics {
		var metricType string
		switch value.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%v", a.serverAddr, metricType, name, value)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("failed to send metric: %v\n", err)
			continue
		}
		resp.Body.Close()
	}
}

func (a *Agent) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	var metrics map[string]interface{}
	for {
		select {
		case <-pollTicker.C:
			metrics = a.collectMetrics()
		case <-reportTicker.C:
			a.sendMetrics(metrics)
		}

	}
}

func main() {
	agent := NewAgent("http://localhost:8080", 2*time.Second, 10*time.Second)
	go agent.Run()
}
