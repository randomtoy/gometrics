package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/randomtoy/gometrics/internal/crypto"
	"github.com/randomtoy/gometrics/internal/model"
	"go.uber.org/zap"
)

type Sender struct {
	log         *zap.SugaredLogger
	config      model.AgentConfig
	metricsChan <-chan []model.Metric
	client      *http.Client
}

func NewSender(log *zap.SugaredLogger, config model.AgentConfig, metricsChan <-chan []model.Metric) *Sender {
	return &Sender{
		log:         log,
		config:      config,
		metricsChan: metricsChan,
		client:      &http.Client{},
	}
}

func (s *Sender) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(s.config.ReportInterval) * time.Second)
	defer ticker.Stop()

	workers := s.config.RateLimit
	if workers < 1 {
		workers = 1
	}

	workerPool := make(chan struct{}, workers)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := <-s.metricsChan
			workerPool <- struct{}{}
			go func(metrics []model.Metric) {
				defer func() { <-workerPool }()
				s.sendMetricsBatch(metrics)
			}(metrics)
		}
	}
}

func (s *Sender) sendMetricsBatch(metrics []model.Metric) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		s.log.Errorf("failed to compress data: %v", err)
		return
	}
	gzipWriter.Close()

	url := fmt.Sprintf("http://%s/updates/", s.config.Addr)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		s.log.Errorf("Can't wrap Request: %v", err)
		return
	}
	if s.config.Key != "" {
		hash := crypto.ComputeHMACSHA256(string(jsonData), s.config.Key)
		req.Header.Set("HashSHA256", hash)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	for attempt := 1; attempt <= 4; attempt++ {
		resp, err := s.client.Do(req)
		if err == nil {
			resp.Body.Close()
			return
		}
		backoff := time.Duration((attempt-1)*2+1) * time.Second
		s.log.Errorf("Can't send metrics %v due to error: %v", backoff, err)
		time.Sleep(backoff)
	}
	s.log.Errorf("failed to send metrics after retries: %v", err)
}
