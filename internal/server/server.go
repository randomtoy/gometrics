package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/randomtoy/gometrics/internal/compress"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/logger"
	"go.uber.org/zap"
)

type Server struct {
	log     *zap.SugaredLogger
	handler *handlers.Handler
}

func NewServer(l *zap.SugaredLogger, h *handlers.Handler) *Server {
	return &Server{
		log:     l,
		handler: h,
	}
}

func (s *Server) Run(addr string) error {
	e := echo.New()

	e.Use(middleware.Gzip())

	e.Use(logger.ResponseLogger(*s.log))
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
	err := e.Start(addr)
	if err != nil {
		return fmt.Errorf("error starting echo: %w", err)
	}
	return nil
}
