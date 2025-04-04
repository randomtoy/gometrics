package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/randomtoy/gometrics/internal/model"
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
	ctx := c.Request().Context()
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

	var metric model.Metric
	metric.ID = path.metricName
	metric.Type = model.MetricType(path.metricType)
	//TODO Return metric
	switch metric.Type {
	case model.Gauge:
		value, err := strconv.ParseFloat(path.metricValue, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		metric.Value = &value
		_, _ = h.store.UpdateMetric(ctx, metric)
	case model.Counter:
		value, err := strconv.ParseInt(path.metricValue, 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintln("Error converting metric"))
		}
		metric.Delta = &value
		_, _ = h.store.UpdateMetric(ctx, metric)
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
	ctx := c.Request().Context()
	metrics, _ := h.store.GetAllMetrics(ctx)

	var result []string
	for _, metric := range metrics {
		result = append(result, fmt.Sprintf("%s: %s (%v)", metric.ID, metric.String(), metric.Type))
	}
	response := strings.Join(result, "\n")

	return c.HTML(http.StatusOK, response)
}

func (h *Handler) HandleMetrics(c echo.Context) error {
	ctx := c.Request().Context()
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
	metric, err := h.store.GetMetric(ctx, path.metricName)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Cant find metric: %s", err))
	}

	return c.String(http.StatusOK, metric.String())
}

func (h *Handler) UpdateMetricJSON(c echo.Context) error {
	ctx := c.Request().Context()
	var metric model.Metric
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

	m, _ := h.store.UpdateMetric(ctx, metric)
	return c.JSON(http.StatusOK, echo.Map{"info": m})
}

func (h *Handler) GetMetricJSON(c echo.Context) error {
	ctx := c.Request().Context()
	var metric model.Metric
	err := c.Bind(&metric)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid JSON"})
	}

	if metric.ID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid metric name"})
	}
	m, err := h.store.GetMetric(ctx, metric.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": fmt.Sprintf("%v", err)})
	}

	return c.JSON(http.StatusOK, m)

}

func (h *Handler) PingDBHandler(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.store.Ping(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "DB is not OK"})
	}
	return c.JSON(http.StatusOK, echo.Map{"info": "DB is OK"})

}

func (h *Handler) BatchHandler(c echo.Context) error {
	ctx := c.Request().Context()
	var metrics []model.Metric

	err := c.Bind(&metrics)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err})
	}
	err = h.store.UpdateMetricBatch(ctx, metrics)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": fmt.Sprintf("%v", err)})
	}
	return c.JSON(http.StatusOK, metrics)
}
