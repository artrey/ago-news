package business

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Service struct {
	pool *pgxpool.Pool
}

type News struct {
	Id      int64
	Title   string
	Text    string
	Image   string
	Created int64
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) CreateNews(ctx context.Context, title, text, image string) (*News, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO news(title, text, image) VALUES($1, $2, $3) RETURNING id, extract(epoch from created)::integer`,
		title, text, image,
	)

	news := News{
		Title: title,
		Text:  text,
		Image: image,
	}

	err := row.Scan(&news.Id, &news.Created)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &news, nil
}

func (s *Service) GetLatestNews(ctx context.Context) ([]*News, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, title, text, image, extract(epoch from created)::integer FROM news ORDER BY created DESC LIMIT 5`,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	latestNews := make([]*News, 0)
	for rows.Next() {
		var news News
		err = rows.Scan(&news.Id, &news.Title, &news.Text, &news.Image, &news.Created)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		latestNews = append(latestNews, &news)
	}

	return latestNews, nil
}
