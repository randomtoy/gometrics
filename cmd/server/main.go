package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/storage"
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
	e := echo.New()
	e.GET("/", s.handler.HandleAllMetrics)
	e.GET("/value/*", s.handler.HandleMetrics)
	e.POST("/update/*", s.handler.HandleUpdate)
	e.Any("/*", func(c echo.Context) error {
		return c.String(http.StatusNotFound, "Page not found")
	})
	err := e.Start(addr)
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

	store := storage.NewInMemoryStorage()
	handler := handlers.NewHandler(store)
	srv := NewServer(handler)

	err := srv.Run(config.Addr)
	if err != nil {
		panic(err)
	}
}
