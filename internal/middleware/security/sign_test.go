package security

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHash(t *testing.T) {
	type args struct {
		correctHash string
		requestHash string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test not equal",
			args: struct {
				correctHash string
				requestHash string
			}{
				correctHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
				requestHash: "4qwrrg582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
			},
			wantErr: true,
		},
		{
			name: "Test equal",
			args: struct {
				correctHash string
				requestHash string
			}{
				correctHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
				requestHash: "43baa3a4500d582a9370c29d7cf44742464942bc910b6ddadd3a7f53bd7e8f68",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := CheckHash(
					tt.args.correctHash,
					tt.args.requestHash,
				); (err != nil) != tt.wantErr {
					t.Errorf("CheckHash() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestHash(t *testing.T) {

	tests := []struct {
		name      string
		data      string
		secretKey []byte
		wantHash  string
	}{
		{
			name:      "Test with right data",
			data:      "lol",
			secretKey: []byte("secret key"),
			wantHash:  "a381f8dd26f6abbd1d5b43d018799bfea10874e1826f033f83084b429e7b75f4",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := Hash(tt.data, tt.secretKey)
				if got != tt.wantHash {
					t.Errorf("Hash() = %v, want %v", got, tt.wantHash)
				}
			},
		)
	}
}

func performRequest(r http.Handler, method, path, body, hash string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Add("HashSHA256", hash)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHashCheckMiddleware_ValidHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	secretKey := "secret"
	router.POST(
		"/test", HashCheckMiddleware(secretKey), func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	body := "test body"
	hash := Hash(body, []byte(secretKey))
	w := performRequest(router, "POST", "/test", body, hash)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHashCheckMiddleware_InvalidHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	secretKey := "secret"
	router.POST(
		"/test", HashCheckMiddleware(secretKey), func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	body := "test body"
	invalidHash := Hash(body+"make it wrong", []byte(secretKey))
	w := performRequest(router, "POST", "/test", body, invalidHash)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHashCheckMiddleware_NoHashProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	secretKey := "secret"
	router.POST(
		"/test", HashCheckMiddleware(secretKey), func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	body := "test body"
	w := performRequest(router, "POST", "/test", body, "")

	assert.Equal(
		t,
		http.StatusOK,
		w.Code,
		"Middleware should allow requests without HashSHA256 header",
	)
}

func TestHashCheckMiddleware_EmptySecretKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(
		"/test", HashCheckMiddleware(""), func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	body := "test body"
	hash := Hash(body, []byte(""))
	w := performRequest(router, "POST", "/test", body, hash)

	assert.Equal(
		t,
		http.StatusOK,
		w.Code,
		"Middleware should not process hash check if secret key is empty",
	)
}
