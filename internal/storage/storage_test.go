package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_UpdateMetric(t *testing.T) {
	store := NewInMemoryStorage()

	t.Run("Update gauge metric", func(t *testing.T) {
		value := float64(123.45)
		store.UpdateGauge("TestGauge", &value)

		metrics := store.GetAllMetrics()
		assert.Contains(t, metrics, "TestGauge")
		assert.Equal(t, &value, metrics["TestGauge"].Value)
	})

	t.Run("Update counter metric", func(t *testing.T) {
		value := int64(10)
		store.UpdateCounter("TestCounter", &value)

		metric, err := store.GetMetric("TestCounter")
		assert.NoError(t, err)
		// assert.Contains(t, metric, "TestCounter")
		assert.Equal(t, &value, metric.Delta)
	})

}

func TestInMemoryStorage_GetMetric(t *testing.T) {
	store := NewInMemoryStorage()

	t.Run("Get gauge metric", func(t *testing.T) {
		value := float64(123.45)
		store.UpdateGauge("TestGauge", &value)
		// assert.NoError(t, err)

		metric, err := store.GetMetric("TestGauge")
		assert.NoError(t, err)

		assert.Equal(t, 123.45, metric.Value)
	})

	t.Run("Get counter metric", func(t *testing.T) {
		counter := int64(10)
		store.UpdateCounter("TestCounter", &counter)

		metric, err := store.GetMetric("TestCounter")
		assert.NoError(t, err)
		assert.Equal(t, &counter, metric.Delta)
	})
	t.Run("Get unknown metric", func(t *testing.T) {
		_, err := store.GetMetric("UnknownName")
		assert.Error(t, err)
	})
}
