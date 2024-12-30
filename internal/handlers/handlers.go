package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/randomtoy/gometrics/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	metricType := storage.MetricType(parts[0])
	name := parts[1]
	valueStr := parts[2]

	var value interface{}
	var err error

	switch metricType {
	case storage.Gauge:
		value, err = strconv.ParseFloat(valueStr, 64)
	case storage.Counter:
		value, err = strconv.ParseInt(valueStr, 10, 64)
	default:
		http.Error(w, "invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	err = h.store.UpdateMetric(metricType, name, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Metric updated")
}

func (h *Handler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusNotFound)
		return
	}
	metrics := h.store.GetAllMetrics()
	for name, metric := range metrics {
		fmt.Fprintf(w, "%s: %v (%s)\n", name, metric.Value, metric.Type)
	}
}
