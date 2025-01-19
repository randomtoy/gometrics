package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_UpdateMetric(t *testing.T) {
	store := NewInMemoryStorage()

	t.Run("Update gauge metric", func(t *testing.T) {
		expect := Metric{
			ID:    "testgauge",
			Value: 123.45,
			Type:  "gauge",
		}
		result, err := store.UpdateMetric("gauge", "TestGauge", "123.45")
		assert.NoError(t, err)
		assert.Equal(t, expect, result)

	})

	t.Run("Update counter metric", func(t *testing.T) {
		expect := Metric{
			ID:    "testcounter",
			Value: int64(15),
			Type:  "counter",
		}
		_, err := store.UpdateMetric("counter", "TestCounter", "10")
		assert.NoError(t, err)

		result, err := store.UpdateMetric("counter", "TestCounter", "5")
		assert.NoError(t, err)

		assert.Equal(t, expect, result)
	})

	t.Run("Invalid gauge value", func(t *testing.T) {
		_, err := store.UpdateMetric("counter", "InvalidGauge", "invalid")
		assert.Error(t, err)
	})

	t.Run("Invalid counter value", func(t *testing.T) {
		_, err := store.UpdateMetric("counter", "InvalidCounter", "invalid")
		assert.Error(t, err)
	})

	t.Run("Invalid metric type", func(t *testing.T) {
		_, err := store.UpdateMetric("Unknown", "UnknownMetric", "unknown")
		assert.Error(t, err)
	})

}

func TestInMemoryStorage_GetMetric(t *testing.T) {
	store := NewInMemoryStorage()

	t.Run("Get gauge metric", func(t *testing.T) {
		_, err := store.UpdateMetric("gauge", "TestGauge", "123.45")
		assert.NoError(t, err)

		metric, err := store.GetMetric("TestGauge")
		assert.NoError(t, err)

		assert.Equal(t, 123.45, metric.Value)
	})

	t.Run("Get counter metric", func(t *testing.T) {
		_, err := store.UpdateMetric("counter", "TestCounter", "10")
		assert.NoError(t, err)

		metric, err := store.GetMetric("TestCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(10), metric.Value)
	})
	t.Run("Get unknown metric", func(t *testing.T) {
		_, err := store.GetMetric("UnknownName")
		assert.Error(t, err)
	})
}
