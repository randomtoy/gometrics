package main

import (
	"github.com/randomtoy/gometrics/internal/agent"
	"github.com/randomtoy/gometrics/internal/config"
	"go.uber.org/zap"
)

const (
	app string = "agent"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	config := config.NewConfig(app)
	a := agent.NewAgent(log.Sugar(), config.Agent)
	a.Run()

}
