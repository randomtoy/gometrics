// storage/storage_test.go
package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_UpdateMetric(t *testing.T) {
	store := NewInMemoryStorage()

	t.Run("Update gauge metric", func(t *testing.T) {
		err := store.UpdateMetric(Gauge, "TestGauge", 123.45)
		assert.NoError(t, err)

		metrics := store.GetAllMetrics()
		assert.Contains(t, metrics, "TestGauge")
		assert.Equal(t, 123.45, metrics["TestGauge"].Value)
	})

	t.Run("Update counter metric", func(t *testing.T) {
		err := store.UpdateMetric(Counter, "TestCounter", int64(10))
		assert.NoError(t, err)

		err = store.UpdateMetric(Counter, "TestCounter", int64(5))
		assert.NoError(t, err)

		metrics := store.GetAllMetrics()
		assert.Contains(t, metrics, "TestCounter")
		assert.Equal(t, int64(15), metrics["TestCounter"].Value)
	})

	t.Run("Invalid gauge value", func(t *testing.T) {
		err := store.UpdateMetric(Gauge, "InvalidGauge", "invalid")
		assert.Error(t, err)
	})

	t.Run("Invalid counter value", func(t *testing.T) {
		err := store.UpdateMetric(Counter, "InvalidCounter", "invalid")
		assert.Error(t, err)
	})
}
