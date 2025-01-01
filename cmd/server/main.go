package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/storage"
)

var enpointAddr string

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

func parseFlags() {
	flag.StringVar(&enpointAddr, "a", "localhost:8080", "endpoint address")
	flag.Parse()
}

func main() {

	parseFlags()

	store := storage.NewInMemoryStorage()
	handler := handlers.NewHandler(store)
	srv := NewServer(handler)

	err := srv.Run(enpointAddr)
	if err != nil {
		panic(err)
	}
}
