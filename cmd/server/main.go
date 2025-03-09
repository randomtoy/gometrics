package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/randomtoy/gometrics/internal/compress"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/logger"
	"github.com/randomtoy/gometrics/internal/model"
	"github.com/randomtoy/gometrics/internal/storage"
	"go.uber.org/zap"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer(handler *handlers.Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Run(addr string) error {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer l.Sync()
	e := echo.New()

	e.Use(middleware.Gzip())

	e.Use(logger.ResponseLogger(*l))
	e.Use(compress.GzipDecompress)

	e.GET("/", s.handler.HandleAllMetrics)
	e.GET("/ping", s.handler.PingDBHandler)
	e.POST("/value/", s.handler.GetMetricJSON)
	e.GET("/value/*", s.handler.HandleMetrics)
	e.POST("/update/", s.handler.UpdateMetricJSON)
	e.POST("/update/*", s.handler.HandleUpdate)
	e.POST("/updates/", s.handler.BatchHandler)

	e.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "Page not found")
	})
	err = e.Start(addr)
	if err != nil {
		return fmt.Errorf("error starting echo: %w", err)
	}
	return nil
}

func parseFlags(config *model.Config) {
	flag.StringVar(&config.DatabaseDSN, "d", "", "PGconnection string")
	flag.StringVar(&config.Addr, "a", "localhost:8080", "endpoint address")
	flag.IntVar(&config.StoreInterval, "i", 10, "Store metric niterval")
	flag.StringVar(&config.FilePath, "f", "", "file path")
	flag.BoolVar(&config.Restore, "r", true, "Restore metrics")

	flag.Parse()
}

func parseEnvironmentFlags(config *model.Config) {
	value, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Addr = value
	}
	si, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		config.StoreInterval, _ = strconv.Atoi(si)
	}
	fsp, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		config.FilePath = fsp
	}
	r, ok := os.LookupEnv("RESTORE")
	if ok {
		config.Restore, _ = strconv.ParseBool(r)
	}
	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		config.DatabaseDSN = dsn
	}
}

func main() {
	var config model.Config
	parseFlags(&config)
	parseEnvironmentFlags(&config)
	l, _ := zap.NewProduction()
	defer l.Sync()

	store, err := storage.NewStorage(l, config)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	handler := handlers.NewHandler(store, handlers.WithLogger(l))

	srv := NewServer(handler)

	err = srv.Run(config.Addr)
	if err != nil {
		panic(err)
	}
}
