package main

import (
	"github.com/randomtoy/gometrics/internal/config"
	"github.com/randomtoy/gometrics/internal/handlers"
	"github.com/randomtoy/gometrics/internal/server"
	"github.com/randomtoy/gometrics/internal/storage"
	"go.uber.org/zap"
)

const (
	app string = "server"
)

func main() {

	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	conf := config.NewConfig(app)

	store, err := storage.NewStorage(l, *conf)
	if err != nil {

		panic(err)
	}
	defer store.Close()

	handler := handlers.NewHandler(store, handlers.WithLogger(l))

	srv := server.NewServer(l.Sugar(), handler)

	err = srv.Run(conf.Server.Addr)
	if err != nil {
		panic(err)
	}
}
