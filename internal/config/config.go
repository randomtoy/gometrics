package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/randomtoy/gometrics/internal/model"
)

func NewConfig(app string) *model.Config {
	var config model.Config
	if app == string(model.AgentApp) {
		parseAgentFlags(&config)
		parseAgentEnvironment(&config)
	}
	if app == string(model.ServerApp) {
		parseServerFlags(&config)
		parseServerEnvironment(&config)
	}
	return &config
}

func parseAgentFlags(config *model.Config) {
	flag.StringVar(&config.Agent.Addr, "a", "localhost:8080", "server address")
	flag.IntVar(&config.Agent.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&config.Agent.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&config.Agent.Key, "k", "", "key")
	flag.IntVar(&config.Agent.RateLimit, "l", 10, "rate limit")

	flag.Parse()
}

func parseAgentEnvironment(config *model.Config) {
	addr, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Agent.Addr = addr
	}
	rep, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		repInt, err := strconv.Atoi(rep)
		if err == nil {
			config.Agent.ReportInterval = repInt
		}
	}
	poll, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		pollInt, err := strconv.Atoi(poll)
		if err == nil {
			config.Agent.PollInterval = pollInt
		}

	}

	key, ok := os.LookupEnv("KEY")
	if ok {
		config.Agent.Key = key
	}
	rate, ok := os.LookupEnv("RATE_LIMIT")
	if ok {
		rateLimit, err := strconv.Atoi(rate)
		if err == nil {
			config.Agent.RateLimit = rateLimit
		}
	}

}

func parseServerFlags(config *model.Config) {
	flag.StringVar(&config.Server.DatabaseDSN, "d", "", "PGconnection string")
	flag.StringVar(&config.Server.Addr, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&config.Server.StoreInterval, "i", 10, "Store metric niterval")
	flag.StringVar(&config.Server.FilePath, "f", "", "file path")
	flag.BoolVar(&config.Server.Restore, "r", true, "Restore metrics")
	flag.StringVar(&config.Server.Key, "k", "", "Key")

	flag.Parse()
}

func parseServerEnvironment(config *model.Config) {
	value, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Server.Addr = value
	}
	si, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		config.Server.StoreInterval, _ = strconv.Atoi(si)
	}
	fsp, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		config.Server.FilePath = fsp
	}
	r, ok := os.LookupEnv("RESTORE")
	if ok {
		config.Server.Restore, _ = strconv.ParseBool(r)
	}
	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		config.Server.DatabaseDSN = dsn
	}
	key, ok := os.LookupEnv("KEY")
	if ok {
		config.Server.Key = key
	}
}
