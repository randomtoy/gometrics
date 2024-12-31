package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/randomtoy/gometrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestHandlers_HandleUpdate(t *testing.T) {
	e := echo.New()
	store := storage.NewInMemoryStorage()
	handler := NewHandler(store)

	t.Run("Valid gauge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/TestGauge/123.45", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleUpdate(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, store.GetAllMetrics(), "TestGauge")
		assert.Equal(t, 123.45, store.GetAllMetrics()["TestGauge"].Value)
	})

	t.Run("Valid counter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/TestCounter/123", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		handler.HandleUpdate(ctx)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, store.GetAllMetrics(), "TestCounter")
		assert.Equal(t, int64(123), store.GetAllMetrics()["TestCounter"].Value)
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
	e := echo.New()
	store := storage.NewInMemoryStorage()
	handler := NewHandler(store)

	store.UpdateMetric(storage.Gauge, "TestGauge", 123.45)
	store.UpdateMetric(storage.Counter, "TestCounter", int64(123))

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
	e := echo.New()
	store := storage.NewInMemoryStorage()
	handler := NewHandler(store)

	store.UpdateMetric(storage.Gauge, "TestGauge", 123.45)
	store.UpdateMetric(storage.Counter, "TestCounter", int64(123))

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
		assert.Contains(t, rec.Body.String(), "123")
	})

	t.Run("Invalid metric type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/unknown/UnknownMetric", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		err := handler.HandleMetrics(ctx)
		assert.Error(t, err)

	})
}
