package main

import (
	"net/http"

	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/storage"
)

type Server struct {
	handler *handlers.Handler
}

func NewServer(handler *handlers.Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", s.handler.HandleUpdate)
	mux.HandleFunc("/metrics", s.handler.HandleMetrics)

	return http.ListenAndServe(addr, mux)
}

func main() {
	store := storage.NewInMemoryStorage()
	handler := handlers.NewHandler(store)
	srv := NewServer(handler)

	err := srv.Run("http://localhost:8080")
	if err != nil {
		panic(err)
	}
}
