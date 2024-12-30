package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MetricStore struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMetricStore() *MetricStore {
	return &MetricStore{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MetricStore) UpdateGauge(name string, value float64) {
	ms.gauges[name] = value
}

func (ms *MetricStore) UpdateCounter(name string, value int64) {
	ms.counters[name] += value
}

func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	result := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (ms *MetricStore) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pathParts := splitPath(r.URL.Path)
		fmt.Printf("pathParts: %+v\n", pathParts)
		log.Printf("pathParts: %+v\n", pathParts)
		log.Default().Printf("pathParts: %+v\n", pathParts)
		if len(pathParts) != 4 || pathParts[0] != "update" {
			http.Error(w, "Invalid URL path", http.StatusNotFound)
			return
		}

		metricType := MetricType(pathParts[1])
		metricName := pathParts[2]
		metricValue := pathParts[3]

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		switch metricType {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			ms.UpdateGauge(metricName, value)
		case Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			ms.UpdateCounter(metricName, value)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
	}

}

func (ms *MetricStore) mainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		res := fmt.Sprintf("metrics: %+v", ms)
		w.Write([]byte(res))
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func main() {
	store := NewMetricStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", store.mainPage)
	mux.HandleFunc("/update/", store.HandleUpdate)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
