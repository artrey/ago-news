package dto

import "github.com/artrey/ago-news/pkg/business"

type NewsDTO struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Text    string `json:"text"`
	Image   string `json:"image"`
	Created int64  `json:"created"`
}

func FromNewsModel(news *business.News) *NewsDTO {
	return &NewsDTO{
		Id:      news.Id,
		Title:   news.Title,
		Text:    news.Text,
		Image:   news.Image,
		Created: news.Created,
	}
}
