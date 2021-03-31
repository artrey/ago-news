package cache

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

type Service struct {
	pool          *redis.Pool
	actionTimeout time.Duration
}

func NewService(redisPool *redis.Pool, actionTimeout time.Duration) *Service {
	return &Service{
		pool:          redisPool,
		actionTimeout: actionTimeout,
	}
}

func (s *Service) Get(ctx context.Context, key string) (bytes []byte, err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	reply, err := redis.DoWithTimeout(conn, s.actionTimeout, "GET", key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	value, err := redis.Bytes(reply, err)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return value, nil
}

func (s *Service) Set(ctx context.Context, key string, bytes []byte) (err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	_, err = redis.DoWithTimeout(conn, s.actionTimeout, "SET", key, bytes)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Service) Delete(ctx context.Context, key string) (err error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	_, err = redis.DoWithTimeout(conn, s.actionTimeout, "DEL", key)
	if err != nil {
		log.Println(err)
	}
	return err
}
