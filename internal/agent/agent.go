package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type Agent struct {
	log       *zap.SugaredLogger
	config    model.AgentConfig
	pollCount int64
}

func NewAgent(l *zap.SugaredLogger, config model.AgentConfig) *Agent {
	return &Agent{
		log:       l,
		config:    config,
		pollCount: 0,
	}
}

func (a *Agent) Run() {
	pollTicker := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	var metrics []model.Metric
	for {
		select {
		case <-pollTicker.C:
			metrics = a.collectMetrics()
		case <-reportTicker.C:
			a.sendMetricsBatch(metrics)
		}
	}
}

func (a *Agent) collectMetrics() []model.Metric {
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

func convertToMetrics(data map[string]interface{}) []model.Metric {
	var metrics []model.Metric
	for key, value := range data {
		metric := model.Metric{
			ID: key,
		}

		switch v := value.(type) {
		case float64:
			metric.Type = model.Gauge
			metric.Value = &v
		case int64:
			metric.Type = model.Counter
			metric.Delta = &v
		}
		metrics = append(metrics, metric)
	}
	return metrics
}

func (a *Agent) sendMetricsBatch(metrics []model.Metric) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		a.log.Errorf("failed to compress data: %v", err)
		return
	}
	gzipWriter.Close()

	url := fmt.Sprintf("http://%s/updates/", a.config.Addr)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		a.log.Errorf("Can't wrap Request: %v", err)
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

		a.log.Errorf("Can't send metrics %v due to error: %v", backoff, err)
		time.Sleep(backoff)
	}
	a.log.Errorf("failed to send metrics after retries: %v", err)
}
