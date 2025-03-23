package collector

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/randomtoy/gometrics/internal/model"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type Collector struct {
	log         *zap.SugaredLogger
	config      model.AgentConfig
	metricsChan chan<- []model.Metric
	pollCount   int64
}

func NewCollector(log *zap.SugaredLogger, config model.AgentConfig, metricsChan chan<- []model.Metric) *Collector {
	return &Collector{
		log:         log,
		config:      config,
		metricsChan: metricsChan,
		pollCount:   0,
	}
}

func (c *Collector) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(c.config.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := c.collectMetrics()
			c.metricsChan <- metrics
		}
	}
}

func (c *Collector) collectMetrics() []model.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.pollCount++
	data := map[string]any{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": float64(memStats.GCCPUFraction),
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),

		"PollCount":   c.pollCount,
		"RandomValue": rand.Float64(),
	}

	vMem, err := mem.VirtualMemory()
	if err == nil {
		data["TotalMemory"] = float64(vMem.Total)
		data["FreeMemory"] = float64(vMem.Free)
	}

	cpuUtil, err := cpu.Percent(0, true)
	if err == nil {
		for i, usage := range cpuUtil {
			key := fmt.Sprintf("CPUutilization%d", i)
			data[key] = usage
		}
	}
	me := convertToMetrics(data)
	c.log.Infof("metric: %#v", me)
	return me
}

func convertToMetrics(data map[string]interface{}) []model.Metric {
	var metrics []model.Metric
	for key, value := range data {
		metric := model.Metric{ID: key}
		switch v := value.(type) {
		case float64:
			metric.Type = model.Gauge
			metric.Value = &v
		case int64:
			metric.Type = model.Counter
			metric.Delta = &v
		}
		metrics = append(metrics, metric)
	}
	return metrics
}
