package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/randomtoy/gometrics/internal/compress"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/logger"
	"github.com/randomtoy/gometrics/internal/storage"
	"go.uber.org/zap"
)

type Config struct {
	Addr string `env:"ADDRESS"`
}

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
	e.POST("/value/", s.handler.GetMetricJSON)
	e.GET("/value/*", s.handler.HandleMetrics)
	e.POST("/update/", s.handler.UpdateMetricJSON)
	e.POST("/update/*", s.handler.HandleUpdate)

	e.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "Page not found")
	})
	err = e.Start(addr)
	if err != nil {
		return fmt.Errorf("error starting echo: %w", err)
	}
	return nil
}

func parseFlags(config *Config) {
	flag.StringVar(&config.Addr, "a", "localhost:8080", "endpoint address")
	flag.Parse()
}

func parseEnvironmentFlags(config *Config) {
	value, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Addr = value
	}
}

func main() {
	config := Config{}
	parseFlags(&config)
	parseEnvironmentFlags(&config)
	l, _ := zap.NewProduction()
	defer l.Sync()

	store := storage.NewInMemoryStorage()

	handler := handlers.NewHandler(store, handlers.WithLogger(l))

	srv := NewServer(handler)

	err := srv.Run(config.Addr)
	if err != nil {
		panic(err)
	}
}
