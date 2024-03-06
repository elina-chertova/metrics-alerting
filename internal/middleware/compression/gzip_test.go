package compression

import (
	gzip "github.com/klauspost/pgzip"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzipHandle(t *testing.T) {
	r := gin.New()
	r.Use(GzipHandle())
	r.GET(
		"/test", func(c *gin.Context) {
			c.String(http.StatusOK, "Hello, Gzip!")
		},
	)

	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("Accept-Encoding", "gzip")
	req1.Header.Set("Content-Type", "text/plain")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf(
			"Expected Content-Encoding to be 'gzip', but got: %s",
			w1.Header().Get("Content-Encoding"),
		)
	}
	req2, _ := http.NewRequest(http.MethodPost, "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Header().Get("Content-Encoding") == "gzip" {
		t.Errorf("Expected Content-Encoding not to be 'gzip', but it was")
	}
}

func TestGzipMiddleware(t *testing.T) {
	// Инициализация Gin с тестовым режимом
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Регистрация middleware и тестового маршрута
	router.Use(GzipHandle())
	router.GET(
		"/test", func(c *gin.Context) {
			c.String(http.StatusOK, "This is a test response")
		},
	)

	// Тестирование с поддержкой Gzip
	t.Run(
		"With Gzip support", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			router.ServeHTTP(w, req)

			// Проверка, что контент сжат
			assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

			// Распаковка тела ответа для проверки содержимого
			reader, err := gzip.NewReader(w.Body)
			assert.NoError(t, err)
			defer reader.Close()

			body, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.True(t, strings.Contains(string(body), "This is a test response"))
		},
	)

	// Тестирование без поддержки Gzip
	t.Run(
		"Without Gzip support", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			router.ServeHTTP(w, req)

			// Проверка, что контент не сжат
			assert.NotEqual(t, "gzip", w.Header().Get("Content-Encoding"))

			body, err := io.ReadAll(w.Body)
			assert.NoError(t, err)
			assert.True(t, strings.Contains(string(body), "This is a test response"))
		},
	)
}
