package compression

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
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
