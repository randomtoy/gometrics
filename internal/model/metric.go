package model

import "fmt"

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Metric struct {
	ID    string     `json:"id"`
	Type  MetricType `json:"type"`
	Value *float64   `json:"value,omitempty"`
	Delta *int64     `json:"delta,omitempty"`
}

func (m Metric) String() string {
	switch m.Type {
	case Gauge:
		return fmt.Sprintf("%v", *m.Value)
	case Counter:
		return fmt.Sprintf("%v", *m.Delta)
	}
	return ""
}

func (m *Metric) Summ(i *int64) *Metric {
	*m.Delta = *m.Delta + *i
	return m
}

func (m *Metric) DerefFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func (m *Metric) DerefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
