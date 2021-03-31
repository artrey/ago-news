package app

import (
	"encoding/json"
	"github.com/artrey/ago-news/cmd/service/app/dto"
	"github.com/artrey/ago-news/pkg/business"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Service struct {
	businessSvc *business.Service
	router      chi.Router
}

func NewService(businessSvc *business.Service, router chi.Router) *Service {
	return &Service{
		businessSvc: businessSvc,
		router:      router,
	}
}

func (s *Service) Init() error {
	s.router.Get("/api/news/latest", s.LatestNews)
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
