package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/artrey/ago-news/cmd/service/app/dto"
	cacheMiddleware "github.com/artrey/ago-news/cmd/service/app/middlewares/cache"
	"github.com/artrey/ago-news/pkg/business"
	"github.com/artrey/ago-news/pkg/cache"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
)

type Service struct {
	businessSvc *business.Service
	cacheSvc    *cache.Service
	router      chi.Router
}

func NewService(businessSvc *business.Service, cacheSvc *cache.Service, router chi.Router) *Service {
	return &Service{
		businessSvc: businessSvc,
		cacheSvc:    cacheSvc,
		router:      router,
	}
}

func (s *Service) Init() error {
	cacheMd := cacheMiddleware.Cache(func(ctx context.Context, path string) ([]byte, error) {
		value, err := s.cacheSvc.Get(ctx, path)
		if err != nil && errors.Is(err, redis.ErrNil) {
			return nil, cacheMiddleware.ErrNotInCache
		}
		return value, err
	}, func(ctx context.Context, path string, data []byte) error {
		return s.cacheSvc.Set(ctx, path, data)
	}, func(writer http.ResponseWriter, data []byte) error {
		writer.Header().Set("Content-Type", "application/json")
		_, err := writer.Write(data)
		if err != nil {
			log.Println(err)
		}
		return err
	})

	s.router.With(cacheMd).Get("/api/news/latest", s.LatestNews)
	s.router.Post("/api/news", s.CreateNews)

	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Service) LatestNews(w http.ResponseWriter, r *http.Request) {
	news, err := s.businessSvc.GetLatestNews(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	data := make([]*dto.NewsDTO, len(news))
	for i, n := range news {
		data[i] = dto.FromNewsModel(n)
	}

	writeJson(w, data, http.StatusOK)
}

func (s *Service) CreateNews(w http.ResponseWriter, r *http.Request) {
	newsDto := dto.NewsDTO{}
	err := json.NewDecoder(r.Body).Decode(&newsDto)
	if err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	news, err := s.businessSvc.CreateNews(r.Context(), newsDto.Title, newsDto.Text, newsDto.Image)
	if err != nil {
		log.Println(err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	data := dto.FromNewsModel(news)

	writeJson(w, data, http.StatusCreated)

	go func() {
		// TODO: add message queue and place task to it
		if err := s.cacheSvc.Delete(context.Background(), "/api/news/latest"); err != nil {
			log.Println(err)
		}
	}()
}

func writeJson(w http.ResponseWriter, data interface{}, code int) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		http.Error(w, "response marshaling failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(body)
	if err != nil {
		log.Println(err)
	}
}
