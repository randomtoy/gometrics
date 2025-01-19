package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/randomtoy/gometrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

type MetricGauge struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Value float64 `json:"value,omitempty"`
	// Delta *int64 `json:"delta,omitempty"` //counter
	// Value *float64 `json:"value,omitempty"` //gauge
}
type MetricCounter struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value int64  `json:"value,omitempty"`
	// Delta *int64 `json:"delta,omitempty"` //counter
	// Value *float64 `json:"value,omitempty"` //gauge
}

func TestHandlers_HandleUpdate(t *testing.T) {
	e := echo.New()
	store := storage.NewInMemoryStorage()
	handler := NewHandler(store)

	t.Run("Valid gauge", func(t *testing.T) {
		reqBody := `{
		"id": "testGauge",
		"type": "gauge",
		"value":"123.45"
		}`
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleUpdate(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response MetricGauge
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		fmt.Print(response)
		assert.Equal(t, "testgauge", response.ID)
		assert.Equal(t, "gauge", response.MType)
		assert.NotNil(t, response.Value)
		assert.Equal(t, 123.45, response.Value)
	})

	t.Run("Valid counter", func(t *testing.T) {
		reqBody := `{
			"id": "testcounter",
			"type": "counter",
			"value":"123"
			}`
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		err := handler.HandleUpdate(ctx)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response MetricCounter
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		fmt.Print(response)
		assert.Equal(t, "testcounter", response.ID)
		assert.Equal(t, "counter", response.MType)
		assert.NotNil(t, response.Value)
		assert.Equal(t, int64(123), response.Value)
	})

	t.Run("Invalid metric type", func(t *testing.T) {
		reqBody := `{
			"id": "testcounter",
			"type": "unknown",
			"value":"123"
			}`
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
}
