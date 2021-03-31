package main

import (
	"context"
	"github.com/artrey/ago-news/cmd/service/app"
	"github.com/artrey/ago-news/pkg/business"
	"github.com/artrey/ago-news/pkg/cache"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	defaultHost        = "0.0.0.0"
	defaultPort        = "9999"
	defaultDSN         = "postgres://app:pass@localhost:5432/db"
	defaultRedisDSN    = "localhost:6379"
	redisMaxIdle       = 3
	redisIdleTimeout   = 240 * time.Second
	redisActionTimeout = 5000 * time.Millisecond
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

	redisDsn, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		redisDsn = defaultRedisDSN
	}

	if exec(net.JoinHostPort(host, port), redisDsn, dsn) != nil {
		os.Exit(1)
	}
}

func exec(serviceAddr, cacheDsn, storageDsn string) error {
	conn, err := pgxpool.Connect(context.Background(), storageDsn)
	if err != nil {
		log.Println(err)
		return err
	}
	defer conn.Close()

	redisPool := &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeout,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", cacheDsn) },
	}

	businessSvc := business.NewService(conn)
	cacheSvc := cache.NewService(redisPool, redisActionTimeout)
	router := chi.NewRouter()
	application := app.NewService(businessSvc, cacheSvc, router)
	err = application.Init()
	if err != nil {
		log.Println(err)
		return err
	}

	server := &http.Server{
		Addr:    serviceAddr,
		Handler: application,
	}
	return server.ListenAndServe()
}
