package cache

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
)

var ErrNotInCache = errors.New("key not found in cache")

type FromCacheFunc func(ctx context.Context, path string) ([]byte, error)
type ToCacheFunc func(ctx context.Context, path string, data []byte) error
type WriteDataFunc func(writer http.ResponseWriter, data []byte) error

type cachedResponseWriter struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func newCachedResponseWriter(writer http.ResponseWriter) *cachedResponseWriter {
	return &cachedResponseWriter{
		ResponseWriter: writer,
		buffer:         new(bytes.Buffer),
	}
}

func (c *cachedResponseWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *cachedResponseWriter) Write(bytes []byte) (int, error) {
	_, err := c.buffer.Write(bytes)
	if err != nil {
		log.Println(err)
	}
	return c.ResponseWriter.Write(bytes)
}

func (c *cachedResponseWriter) WriteHeader(statusCode int) {
	c.ResponseWriter.WriteHeader(statusCode)
}

func Cache(
	fromCache FromCacheFunc, toCache ToCacheFunc, writeData WriteDataFunc,
) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			data, err := fromCache(request.Context(), request.RequestURI)
			if err == nil {
				log.Printf("Get from cache: %s", request.RequestURI)
				err = writeData(writer, data)
				if err != nil {
					log.Println(err)
				}
				return
			}

			if !errors.Is(err, ErrNotInCache) {
				log.Println(err)
			}

			cachedWriter := newCachedResponseWriter(writer)
			handler.ServeHTTP(cachedWriter, request)

			go func() {
				err = toCache(context.Background(), request.RequestURI, cachedWriter.buffer.Bytes())
				if err != nil {
					log.Println(err)
				}
			}()
		})
	}
}
