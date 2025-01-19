package handlers

import (
	"fmt"
	"net/http"

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
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value,omitempty"`
}

type Handler struct {
	store storage.Storage
	log   *zap.Logger
}

func NewHandler(store storage.Storage) *Handler {
	logger := zap.NewNop()
	h := &Handler{
		store: store,
		log:   logger,
	}

	return h
}

func (h *Handler) HandleUpdate(c echo.Context) error {

	var metric Metrics
	err := c.Bind(&metric)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Body"})
	}

	if metric.ID == "" || metric.Value == "" {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	m, err := h.store.UpdateMetric(string(metric.MType), metric.ID, metric.Value)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Cant update metric : %s", err)})
	}

	return c.JSON(http.StatusOK, m)
}

func (h *Handler) HandleAllMetrics(c echo.Context) error {
	metrics := h.store.GetAllMetrics()

	return c.JSON(http.StatusOK, metrics)
}

func (h *Handler) HandleMetrics(c echo.Context) error {
	var metric Metrics
	err := c.Bind(&metric)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Body"})
	}

	if metric.ID == "" {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "metric not found"})
	}

	m, err := h.store.GetMetric(metric.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Cant find metric: %s", err)})
	}

	return c.JSON(http.StatusOK, m)
}
