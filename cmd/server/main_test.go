package main

import (
	"net/http"
	"testing"
)

func TestMetricStore_HandleUpdate(t *testing.T) {
	type fields struct {
		gauges   map[string]float64
		counters map[string]int64
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MetricStore{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}
			ms.HandleUpdate(tt.args.w, tt.args.r)
		})
	}
}
