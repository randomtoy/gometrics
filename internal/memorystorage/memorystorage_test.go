package memorystorage

import (
	"context"
	"testing"

	"github.com/randomtoy/gometrics/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInMemoryStorage_UpdateMetric(t *testing.T) {
	ctx := context.Background()
	l := zap.NewNop().Sugar()
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
		store.UpdateMetric(ctx, gaugeMetric)

		metrics, _ := store.GetAllMetrics(ctx)
		assert.Contains(t, metrics, "TestGauge")
		assert.Equal(t, &gaugeValue, metrics["TestGauge"].Value)
	})

	t.Run("Update counter metric", func(t *testing.T) {

		store.UpdateMetric(ctx, counterMetric)

		metric, err := store.GetMetric(ctx, "TestCounter")
		assert.NoError(t, err)
		// assert.Contains(t, metric, "TestCounter")
		assert.Equal(t, &counterValue, metric.Delta)
	})

}

func TestInMemoryStorage_GetMetric(t *testing.T) {
	ctx := context.Background()
	l := zap.NewNop().Sugar()
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

		store.UpdateMetric(ctx, gaugeMetric)
		// assert.NoError(t, err)

		metric, err := store.GetMetric(ctx, "TestGauge")
		assert.NoError(t, err)

		assert.Equal(t, &gaugeValue, metric.Value)
	})

	t.Run("Get counter metric", func(t *testing.T) {

		store.UpdateMetric(ctx, counterMetric)

		metric, err := store.GetMetric(ctx, "TestCounter")
		assert.NoError(t, err)
		assert.Equal(t, &counterValue, metric.Delta)
	})
	t.Run("Get unknown metric", func(t *testing.T) {
		_, err := store.GetMetric(ctx, "UnknownName")
		assert.Error(t, err)
	})
}
