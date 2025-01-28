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

	var metric storage.Metric
	metric.ID = path.metricName
	metric.Type = storage.MetricType(path.metricType)
	//TODO Return metric
	switch metric.Type {
	case storage.Gauge:
		value, err := strconv.ParseFloat(path.metricValue, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		metric.Value = &value
		_ = h.store.UpdateMetric(metric)
	case storage.Counter:
		value, err := strconv.ParseInt(path.metricValue, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		metric.Delta = &value
		_ = h.store.UpdateMetric(metric)
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
		return ""
	}
	return parts[index]
}

func (h *Handler) HandleAllMetrics(c echo.Context) error {
	metrics := h.store.GetAllMetrics()

	var result []string
	for _, metric := range metrics {
		result = append(result, fmt.Sprintf("%s: %s (%v)", metric.ID, metric.String(), metric.Type))
	}
	response := strings.Join(result, "\n")

	return c.HTML(http.StatusOK, response)
}

func (h *Handler) HandleMetrics(c echo.Context) error {
	path := trimPath(c.Request().URL.Path)

	if path.action != string(ActionValue) {
		return c.String(http.StatusNotFound, fmt.Sprintln("Action not found"))
	}

	if path.metricName == "" {
		return c.String(http.StatusNotFound, fmt.Sprintln("Cant find metric name"))
	}
	if path.metricType != "gauge" && path.metricType != "counter" {
		return c.String(http.StatusBadRequest, fmt.Sprintf("invalid metric type: %v", path.metricType))
	}
	metric, err := h.store.GetMetric(path.metricName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Cant find metric: %s", err))
	}

	return c.String(http.StatusOK, metric.String())
}

func (h *Handler) UpdateMetricJSON(c echo.Context) error {
	var metric storage.Metric
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

	m := h.store.UpdateMetric(metric)
	return c.JSON(http.StatusOK, echo.Map{"info": m})
}

func (h *Handler) GetMetricJSON(c echo.Context) error {
	var metric storage.Metric
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
