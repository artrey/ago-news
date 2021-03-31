package main

import (
	"context"
	"github.com/artrey/ago-news/cmd/service/app"
	"github.com/artrey/ago-news/pkg/business"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "9999"
	defaultDSN  = "postgres://app:pass@localhost:5432/db"
)

func main() {
	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = defaultHost
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		dsn = defaultDSN
	}

	if exec(net.JoinHostPort(host, port), dsn) != nil {
		os.Exit(1)
	}
}

func exec(addr, dsn string) error {
	conn, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer conn.Close()

	businessSvc := business.NewService(conn)
	router := chi.NewRouter()
	application := app.NewService(businessSvc, router)
	err = application.Init()
	if err != nil {
		log.Println(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}
