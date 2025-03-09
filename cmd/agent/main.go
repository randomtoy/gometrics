package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/labstack/gommon/log"
)

type Type string

const (
	Gauge   Type = "gauge"
	Counter Type = "counter"
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

type Metric struct {
	Type  Type
	ID    string
	Delta *int64
	Value *float64
}

func convertToMetrics(data map[string]interface{}) []Metric {
	var metrics []Metric
	for key, value := range data {
		metric := Metric{
			ID: key,
		}

		switch v := value.(type) {
		case float64:
			metric.Type = Gauge
			metric.Value = &v
		case int64:
			metric.Type = Counter
			metric.Delta = &v
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func NewAgent(serverAddr string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		serverAddr:     serverAddr,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		pollCount:      0,
	}
}

func (a *Agent) collectMetrics() []Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.pollCount++
	data := map[string]interface{}{
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
	metrics := convertToMetrics(data)
	return metrics
}

func (a *Agent) sendMetricsBatch(metrics []Metric) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	fmt.Printf("%#v", jsonData)
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		fmt.Printf("failed to compress data: %v", err)
		return
	}
	gzipWriter.Close()

	url := fmt.Sprintf("%s/updates/", a.serverAddr)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		log.Errorf("Can't wrap Request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	for attempt := 1; attempt <= 4; attempt++ {
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			resp.Body.Close()
			return
		}
		backoff := time.Duration((attempt-1)*2+1) * time.Second

		fmt.Printf("Can't send metrics %v due to error: %v", backoff, err)
		time.Sleep(backoff)
	}
	fmt.Errorf("failed to send metrics after retries: %w", err)
	return

}

func (a *Agent) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	var metrics []Metric
	for {
		select {
		case <-pollTicker.C:
			metrics = a.collectMetrics()
		case <-reportTicker.C:
			a.sendMetricsBatch(metrics)
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
