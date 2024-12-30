package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type MetricType string

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

type Metric struct {
	Type  MetricType
	Name  string
	Value interface{}
}

type MetricStore struct {
	metrics   map[string]Metric
	PollCount int64
}

func NewMetricStore() *MetricStore {
	return &MetricStore{
		metrics: make(map[string]Metric),
	}
}

func (ms *MetricStore) UpdateGauge(name string, value float64) {
	ms.metrics[name] = Metric{
		Type:  MetricTypeGauge,
		Name:  name,
		Value: value,
	}
}

func (ms *MetricStore) UpdateCounter(name string, value int64) {

	ms.metrics[name] = Metric{
		Type:  MetricTypeCounter,
		Name:  name,
		Value: value,
	}

}

func (ms *MetricStore) CollectRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ms.UpdateGauge("Alloc", float64(memStats.Alloc))
	ms.UpdateGauge("BuckHashSys", float64(memStats.BuckHashSys))
	ms.UpdateGauge("Frees", float64(memStats.Frees))
	ms.UpdateGauge("GCCPUFraction", memStats.GCCPUFraction)
	ms.UpdateGauge("GCSys", float64(memStats.GCSys))
	ms.UpdateGauge("HeapAlloc", float64(memStats.HeapAlloc))
	ms.UpdateGauge("HeapIdle", float64(memStats.HeapIdle))
	ms.UpdateGauge("HeapInuse", float64(memStats.HeapInuse))
	ms.UpdateGauge("HeapObjects", float64(memStats.HeapObjects))
	ms.UpdateGauge("HeapReleased", float64(memStats.HeapReleased))
	ms.UpdateGauge("HeapSys", float64(memStats.HeapSys))
	ms.UpdateGauge("LastGC", float64(memStats.LastGC))
	ms.UpdateGauge("Lookups", float64(memStats.Lookups))
	ms.UpdateGauge("MCacheInuse", float64(memStats.MCacheInuse))
	ms.UpdateGauge("MCacheSys", float64(memStats.MCacheSys))
	ms.UpdateGauge("MSpanInuse", float64(memStats.MSpanInuse))
	ms.UpdateGauge("MSpanSys", float64(memStats.MSpanSys))
	ms.UpdateGauge("Mallocs", float64(memStats.Mallocs))
	ms.UpdateGauge("NextGC", float64(memStats.NextGC))
	ms.UpdateGauge("NumForcedGC", float64(memStats.NumForcedGC))
	ms.UpdateGauge("NumGC", float64(memStats.NumGC))
	ms.UpdateGauge("OtherSys", float64(memStats.OtherSys))
	ms.UpdateGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	ms.UpdateGauge("StackInuse", float64(memStats.StackInuse))
	ms.UpdateGauge("StackSys", float64(memStats.StackSys))
	ms.UpdateGauge("Sys", float64(memStats.Sys))
	ms.UpdateGauge("TotalAlloc", float64(memStats.TotalAlloc))

	ms.UpdateCounter("PollCount", 1)
	ms.UpdateGauge("RandomValue", rand.Float64())

}
func (ms *MetricStore) SendMetrics(serverAddr string) {

	client := &http.Client{}
	for _, metric := range ms.metrics {
		url := serverAddr + "/update/" + string(metric.Type) + "/" + metric.Name + "/" + formatMetricValue(metric.Value)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			log.Printf("Failed to create request for metric %s: %v", metric.Name, err)
			continue
		}
		req.Header.Set("Content-Type", "text/plain")
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send metric %s: %v", metric.Name, err)
			continue
		}
		resp.Body.Close()
	}
}

func formatMetricValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return ""
	}
}

func main() {
	store := NewMetricStore()
	serverAddr := "http://localhost:8080"

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	go func() {
		for {
			store.CollectRuntimeMetrics()
			time.Sleep(pollInterval)
		}
	}()

	go func() {
		for {
			store.SendMetrics(serverAddr)
			time.Sleep(reportInterval)
		}
	}()

	select {}
}
