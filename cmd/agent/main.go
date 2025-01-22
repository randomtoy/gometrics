package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	pollInterval   int    `env:"POLL_INTERVAL"`
}

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
		pollCount:      0,
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
		"GCCPUFraction": float64(memStats.GCCPUFraction),
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
		var data map[string]interface{}
		switch value.(type) {
		case float64:

			metricType = "gauge"
			data = map[string]interface{}{
				"id":    name,
				"type":  metricType,
				"value": value,
			}
		case int64:
			fmt.Printf("Key '%s' is of type int64 with value %d\n", name, value)
			metricType = "counter"
			data = map[string]interface{}{
				"id":    name,
				"type":  metricType,
				"delta": value,
			}
		default:
			continue
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			continue
		}
		url := fmt.Sprintf("%s/update/", a.serverAddr)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")

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
	config := Config{}
	parseFlags(&config)
	parseEnvironmentFlags(&config)

	agent := NewAgent("http://"+config.Addr, time.Duration(config.pollInterval)*time.Second, time.Duration(config.reportInterval)*time.Second)
	agent.Run()
}

func parseFlags(config *Config) {
	flag.StringVar(&config.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&config.reportInterval, "r", 10, "report interval")
	flag.IntVar(&config.pollInterval, "p", 2, "poll interval")

	flag.Parse()
}

func parseEnvironmentFlags(config *Config) {
	addr, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Addr = addr
	}
	rep, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		repInt, err := strconv.Atoi(rep)
		if err == nil {
			config.reportInterval = repInt
		}
	}
	poll, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		pollInt, err := strconv.Atoi(poll)
		if err == nil {
			config.pollInterval = pollInt
		}

	}

}
