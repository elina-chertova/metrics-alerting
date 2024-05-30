package subnet

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedIPMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	trustedIP := "192.168.1.1"

	router.Use(TrustedIPMiddleware(trustedIP))
	router.GET(
		"/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		},
	)

	t.Run(
		"Access granted for trusted IP", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Real-IP", trustedIP)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "access granted")
		},
	)

	t.Run(
		"Access denied for untrusted IP", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Real-IP", "192.168.1.2")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.Contains(t, w.Body.String(), ErrUntrustedSubnet.Error())
		},
	)
}
