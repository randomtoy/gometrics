package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/randomtoy/gometrics/internal/storage"
)

type HandlerAction string

const (
	ActionUpdate HandlerAction = "update"
	ActionValue  HandlerAction = "value"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Handler struct {
	store storage.Storage
	log   *zap.Logger
}

type pathParts struct {
	action      string
	metricType  string
	metricName  string
	metricValue string
}

type Option func(h *Handler)

func NewHandler(store storage.Storage, opts ...Option) *Handler {
	logger := zap.NewNop()
	h := &Handler{
		store: store,
		log:   logger,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}
func WithLogger(l *zap.Logger) Option {
	return func(h *Handler) {
		h.log = l
	}
}

func (h *Handler) HandleUpdate(c echo.Context) error {

	path := trimPath(c.Request().URL.Path)

	// Not sure that is reasonable check, because echo shouldnt routing
	// to this handler anythnig except ActionUpdate
	if path.action != string(ActionUpdate) {
		return c.String(http.StatusNotFound, fmt.Sprintln("Action not found"))
	}

	if path.metricName == "" {
		return c.String(http.StatusNotFound, fmt.Sprintln("Cant find metric name"))
	}
	// Lets check if value exist and return error if not
	if path.metricValue == "" {
		return c.String(http.StatusBadRequest, fmt.Sprintln("Incorrect Value"))

	}
	//TODO Return metric
	switch storage.MetricType(path.metricType) {
	case storage.Gauge:
		value, err := strconv.ParseFloat(path.metricValue, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		h.store.UpdateGauge(path.metricName, &value)
	case storage.Counter:
		value, err := strconv.ParseInt(path.metricValue, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		h.store.UpdateCounter(path.metricName, &value)
	default:
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid metric type: %s", path.metricType))
	}

	return c.String(http.StatusOK, fmt.Sprintln("Metric Updated"))
}

func trimPath(path string) pathParts {
	var paths pathParts
	parts := strings.Split(path, "/")

	paths.action = getElement(parts, 1)
	paths.metricType = getElement(parts, 2)
	paths.metricName = getElement(parts, 3)
	paths.metricValue = getElement(parts, 4)

	return paths

}

func getElement(parts []string, index int) string {
	if index < 0 || index >= len(parts) {
		// return "", fmt.Errorf("index %d out of range (length: %d)", index, len(parts))
		return ""
	}
	return parts[index]
}

func (h *Handler) HandleAllMetrics(c echo.Context) error {
	metrics := h.store.GetAllMetrics()

	var result string
	for name, metric := range metrics {
		result += fmt.Sprintf("%s: %v (%s)\n", name, metric.Value, metric.Type)
	}

	return c.HTML(http.StatusOK, result)
}

func (h *Handler) HandleMetrics(c echo.Context) error {
	path := trimPath(c.Request().URL.Path)

	if path.action != string(ActionValue) {
		return c.String(http.StatusNotFound, fmt.Sprintln("Action not found"))
	}

	if path.metricName == "" {
		return c.String(http.StatusNotFound, fmt.Sprintln("Cant find metric name"))
	}
	metric, err := h.store.GetMetric(path.metricName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Cant find metric: %s", err))
	}
	switch storage.MetricType(path.metricType) {
	case storage.Gauge:
		return c.String(http.StatusOK, fmt.Sprintf("%v", *metric.Value))
	case storage.Counter:
		return c.String(http.StatusOK, fmt.Sprintf("%v", *metric.Delta))
	default:
		return c.String(http.StatusBadRequest, fmt.Sprintf("Unknown metric type: %s", path.metricType))
	}
}

func (h *Handler) UpdateMetricJSON(c echo.Context) error {
	var metric Metrics
	err := c.Bind(&metric)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON"})
	}

	if metric.ID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid metric name"})
	}

	if metric.Delta == nil && metric.Value == nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Empty Value"})
	}

	// TODO Return Metric
	var m storage.Metric
	switch storage.MetricType(metric.MType) {
	case storage.Gauge:
		m = h.store.UpdateGauge(metric.ID, metric.Value)
	case storage.Counter:
		m = h.store.UpdateCounter(metric.ID, metric.Delta)
	default:
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Invalid metric type"})
	}

	return c.JSON(http.StatusOK, echo.Map{"info": m})
}

func (h *Handler) GetMetricJSON(c echo.Context) error {
	var metric Metrics
	err := c.Bind(&metric)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON"})
	}

	if metric.ID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid metric name"})
	}
	m, err := h.store.GetMetric(metric.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Metric not found"})
	}

	return c.JSON(http.StatusOK, m)

}
