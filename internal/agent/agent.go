package agent

import (
	"context"
	"sync"

	"github.com/randomtoy/gometrics/internal/collector"
	"github.com/randomtoy/gometrics/internal/model"
	"github.com/randomtoy/gometrics/internal/sender"
	"go.uber.org/zap"
)

type Agent struct {
	log    *zap.SugaredLogger
	config model.AgentConfig
}

func NewAgent(l *zap.SugaredLogger, config model.AgentConfig) *Agent {
	return &Agent{
		log:    l,
		config: config,
	}
}

func (a *Agent) Run() {

	metricsChan := make(chan []model.Metric, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	collector := collector.NewCollector(a.log, a.config, metricsChan)
	sender := sender.NewSender(a.log, a.config, metricsChan)
	wg.Add(2)

	go collector.Run(ctx, &wg)
	go sender.Run(ctx, &wg)
	wg.Wait()

}
