package memorystorage

import (
	"testing"

	"github.com/randomtoy/gometrics/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInMemoryStorage_UpdateMetric(t *testing.T) {
	l := zap.NewNop()
	store := NewInMemoryStorage(l, "")

	counterValue := int64(10)
	counterMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestCounter",
		Delta: &counterValue,
	}
	gaugeValue := float64(123.45)
	gaugeMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestGauge",
		Value: &gaugeValue,
	}

	t.Run("Update gauge metric", func(t *testing.T) {
		store.UpdateMetric(gaugeMetric)

		metrics, _ := store.GetAllMetrics()
		assert.Contains(t, metrics, "TestGauge")
		assert.Equal(t, &gaugeValue, metrics["TestGauge"].Value)
	})

	t.Run("Update counter metric", func(t *testing.T) {

		store.UpdateMetric(counterMetric)

		metric, err := store.GetMetric("TestCounter")
		assert.NoError(t, err)
		// assert.Contains(t, metric, "TestCounter")
		assert.Equal(t, &counterValue, metric.Delta)
	})

}

func TestInMemoryStorage_GetMetric(t *testing.T) {
	l := zap.NewNop()
	store := NewInMemoryStorage(l, "")

	counterValue := int64(10)
	counterMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestCounter",
		Delta: &counterValue,
	}
	gaugeValue := float64(123.45)
	gaugeMetric := model.Metric{
		Type:  model.Gauge,
		ID:    "TestGauge",
		Value: &gaugeValue,
	}

	t.Run("Get gauge metric", func(t *testing.T) {

		store.UpdateMetric(gaugeMetric)
		// assert.NoError(t, err)

		metric, err := store.GetMetric("TestGauge")
		assert.NoError(t, err)

		assert.Equal(t, &gaugeValue, metric.Value)
	})

	t.Run("Get counter metric", func(t *testing.T) {

		store.UpdateMetric(counterMetric)

		metric, err := store.GetMetric("TestCounter")
		assert.NoError(t, err)
		assert.Equal(t, &counterValue, metric.Delta)
	})
	t.Run("Get unknown metric", func(t *testing.T) {
		_, err := store.GetMetric("UnknownName")
		assert.Error(t, err)
	})
}
