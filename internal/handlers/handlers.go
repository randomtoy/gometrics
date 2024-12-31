package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/randomtoy/gometrics/internal/storage"
)

type HandlerAction string

const (
	ActionUpdate HandlerAction = "update"
	ActionValue  HandlerAction = "value"
)

type Handler struct {
	store storage.Storage
}

type pathParts struct {
	action      string
	metricType  string
	metricName  string
	metricValue string
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{store: store}
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

	var value interface{}
	var err error

	switch storage.MetricType(path.metricType) {
	case storage.Gauge:

		value, err = strconv.ParseFloat(path.metricValue, 64)
	case storage.Counter:
		value, err = strconv.ParseInt(path.metricValue, 10, 64)
	default:
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid metric type: %s", path.metricType))
	}

	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid value: %s", err))
	}

	err = h.store.UpdateMetric(storage.MetricType(path.metricType), path.metricName, value)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Cant update metric : %s", err))
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

	return c.String(http.StatusOK, result)
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
		return c.String(http.StatusOK, fmt.Sprintf("%f", metric.Value.(float64)))
	case storage.Counter:
		return c.String(http.StatusOK, fmt.Sprintf("%d", metric.Value.(int64)))
	default:
		return c.String(http.StatusNotFound, fmt.Sprintf("Unknown metric type: %s", path.metricType))
	}
}
