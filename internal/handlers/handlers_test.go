package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/randomtoy/gometrics/internal/model"
	"github.com/randomtoy/gometrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandlers_HandleUpdate(t *testing.T) {
	l := zap.NewNop()
	config := model.Config{
		Server: model.ServerConfig{
			Restore: false,
		},
	}
	e := echo.New()
	store, err := storage.NewStorage(l, config)
	assert.NoError(t, err)
	handler := NewHandler(store)

	counterValue := int64(10)
	counterMetric := model.Metric{
		Type:  model.Counter,
		ID:    "TestCounter",
		Delta: &counterValue,
	}
	gaugeValue := float64(123.45)
	gaugeMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestGauge",
		Value: &gaugeValue,
	}

	t.Run("Valid gauge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/TestGauge/123.45", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleUpdate(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		am, err := store.GetAllMetrics()
		assert.NoError(t, err)
		assert.Contains(t, am, gaugeMetric.ID)
		assert.Equal(t, &gaugeValue, am[gaugeMetric.ID].Value)
	})

	t.Run("Valid counter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/TestCounter/10", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		handler.HandleUpdate(ctx)

		assert.Equal(t, http.StatusOK, rec.Code)
		am, err := store.GetAllMetrics()
		assert.NoError(t, err)
		assert.Contains(t, am, counterMetric.ID)
		fmt.Printf("%#v", am)
		assert.Equal(t, &counterValue, am[counterMetric.ID].Delta)

	})

	t.Run("Invalid metric type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/unknown/UnknownMetric/10", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		handler.HandleUpdate(ctx)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})

	t.Run("Invalid value", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/TestGauge/string", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		handler.HandleUpdate(ctx)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Metric without name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		handler.HandleUpdate(ctx)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestHandler_HandleAllMetrics(t *testing.T) {
	l := zap.NewNop()
	config := model.Config{
		Server: model.ServerConfig{
			Restore: false,
		},
	}
	e := echo.New()
	store, err := storage.NewStorage(l, config)
	assert.NoError(t, err)
	handler := NewHandler(store)
	counterValue := int64(10)
	counterMetric := model.Metric{
		Type:  model.Counter,
		ID:    "TestCounter",
		Delta: &counterValue,
	}
	gaugeValue := float64(123.45)
	gaugeMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestGauge",
		Value: &gaugeValue,
	}
	store.UpdateMetric(counterMetric)
	store.UpdateMetric(gaugeMetric)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)
	handler.HandleAllMetrics(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, _ := io.ReadAll(rec.Body)
	assert.Contains(t, string(body), "TestGauge")
	assert.Contains(t, string(body), "TestCounter")
}

func TestHandler_HandleGetMetric(t *testing.T) {
	l := zap.NewNop()
	config := model.Config{
		Server: model.ServerConfig{
			Restore: false,
		},
	}
	e := echo.New()
	store, err := storage.NewStorage(l, config)
	assert.NoError(t, err)
	handler := NewHandler(store)
	counterValue := int64(10)
	counterMetric := model.Metric{
		Type:  model.Counter,
		ID:    "TestCounter",
		Delta: &counterValue,
	}
	gaugeValue := float64(123.45)
	gaugeMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestGauge",
		Value: &gaugeValue,
	}
	store.UpdateMetric(counterMetric)
	store.UpdateMetric(gaugeMetric)

	t.Run("Valid gauge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/gauge/TestGauge", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleMetrics(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "123.45")
	})

	t.Run("Valid counter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/counter/TestCounter", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleMetrics(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "10")
	})

	t.Run("Invalid metric name", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/value/counter/UnknownMetric", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		handler.HandleMetrics(ctx)
		assert.Equal(t, http.StatusNotFound, rec.Code)

	})
	t.Run("Invalid metric type", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/value/unknown/TestCounter", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		handler.HandleMetrics(ctx)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

	})
	t.Run("Empty metric Name", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/gauge//", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		handler.HandleMetrics(ctx)
		assert.Equal(t, http.StatusNotFound, rec.Code)

	})

	t.Run("Diff cases metric test", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/gauge/TestGauge", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := handler.HandleMetrics(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "123.45")
	})

}
