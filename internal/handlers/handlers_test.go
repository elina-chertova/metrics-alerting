package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"

	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected int
	}{
		{
			name:     "Valid Gauge Metric",
			path:     "/update/gauge/metric1/42.5",
			method:   http.MethodPost,
			expected: http.StatusOK,
		},
		{
			name:     "Valid Counter Metric",
			path:     "/update/counter/metric2/10",
			method:   http.MethodPost,
			expected: http.StatusOK,
		},
		{
			name:     "Invalid Metric Type",
			path:     "/update/invalid/metric3/5",
			method:   http.MethodPost,
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid Metric Value",
			path:     "/update/counter/metric2/i",
			method:   http.MethodPost,
			expected: http.StatusBadRequest,
		},
		{
			name:     "Invalid Metric Value",
			path:     "/update/counter/2",
			method:   http.MethodPost,
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				st := filememory.NewMemStorage(false, nil)
				h := NewHandler(st)
				router.POST(
					"/update/:metricType/:metricName/:metricValue",
					h.MetricsTextPlainHandler(),
				)

				request := httptest.NewRequest(tt.method, tt.path, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, request)
				result := w.Result()
				defer result.Body.Close()
				assert.Equal(t, tt.expected, result.StatusCode)
			},
		)
	}
}

func TestGetMetricsTextPlainHandler(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		postPath string
		expected int
	}{
		{
			name:     "Valid Gauge Metric",
			path:     "/value/gauge/metric1",
			postPath: "/update/gauge/metric1/42.5",
			expected: http.StatusOK,
		},
		{
			name:     "Valid Counter Metric",
			path:     "/value/counter/metric2",
			postPath: "/update/counter/metric2/10",
			expected: http.StatusOK,
		},
		{
			name:     "Invalid Metric Type",
			path:     "/value/invalid/metric3",
			expected: http.StatusBadRequest,
		},
		{
			name:     "Metric Not Found",
			path:     "/value/gauge/nonexistent",
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				st := filememory.NewMemStorage(false, nil)
				h := NewHandler(st)

				if tt.expected == http.StatusOK {
					router.POST(
						"/update/:metricType/:metricName/:metricValue",
						h.MetricsTextPlainHandler(),
					)
					request := httptest.NewRequest(http.MethodPost, tt.postPath, nil)
					w := httptest.NewRecorder()
					router.ServeHTTP(w, request)
				}

				router.GET("/value/:metricType/:metricName", h.GetMetricsTextPlainHandler(""))
				request := httptest.NewRequest(http.MethodGet, tt.path, nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, request)
				result := w.Result()
				defer result.Body.Close()
				assert.Equal(t, tt.expected, result.StatusCode)
			},
		)
	}
}

func TestMetricsJSONHandler(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected int
	}{
		{
			name:     "Valid Gauge Metric",
			payload:  `{"id":"metric1", "type":"gauge", "value":42.5}`,
			expected: http.StatusOK,
		},
		{
			name:     "Valid Counter Metric",
			payload:  `{"id":"metric2", "type":"counter", "delta":10}`,
			expected: http.StatusOK,
		},
		{
			name:     "Invalid JSON",
			payload:  `{"id":"metric3", "type":"gauge"}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				router := gin.Default()
				st := filememory.NewMemStorage(false, nil)
				h := NewHandler(st)
				router.POST("/update/", h.MetricsJSONHandler(""))

				request := httptest.NewRequest(
					http.MethodPost,
					"/update/",
					strings.NewReader(tt.payload),
				)
				request.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, request)
				result := w.Result()
				defer result.Body.Close()
				assert.Equal(t, tt.expected, result.StatusCode)
			},
		)
	}
}

func TestGetMetricsJSONHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	serverConfig := config.NewServer()
	memStorage := filememory.NewMemStorage(true, serverConfig)
	h := NewHandler(memStorage)

	router.POST("/value/", h.GetMetricsJSONHandler("your-secret-key"))

	tests := []struct {
		name           string
		requestBody    []byte
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Invalid Metric Type",
			requestBody:    []byte(`{"id": "metric2", "type": "invalid", "delta": 5}`),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported metric type",
		},
		{
			name:           "Missing Metric type",
			requestBody:    []byte(`{"id": "metric33", "value": 5}`),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "unsupported metric type",
		},
		{
			name:           "Invalid JSON Format",
			requestBody:    []byte(`{"id": "metric3", "type": "counter", "delta": "invalid"}`),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "json: cannot unmarshal string into Go struct field Metric.delta of type int64",
		},
		{
			name:           "Valid Counter Metric",
			requestBody:    []byte(`{"id": "metric1", "type": "counter", "value": 10}`),
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Valid Gauge Metric",
			requestBody:    []byte(`{"id": "metric2", "type": "gauge", "value": 5.5}`),
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(tc.requestBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				responseBody, err := io.ReadAll(w.Body)
				assert.NoError(t, err, "Reading response body should not produce an error")

				assert.Equal(
					t,
					tc.expectedStatus,
					w.Code,
					"Expected and actual status codes should match",
				)

				if w.Code != tc.expectedStatus {
					t.Errorf(
						"Expected status code %d, got %d, response: %s",
						tc.expectedStatus,
						w.Code,
						responseBody,
					)
				}
				if tc.expectedBody != "" {
					assert.Contains(
						t,
						string(responseBody),
						tc.expectedBody,
						"Response body should contain the expected message",
					)
				}
			},
		)
	}
}

func Example_updateBatchMetrics() {
	router := gin.Default()

	ss := &config.Server{
		FlagAddress:     "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "tmp/metrics-db.json",
		FlagRestore:     true,
		DatabaseDSN:     "",
		SecretKey:       "secret-key",
	}

	s := filememory.NewMemStorage(true, ss)
	h := NewHandler(s)

	router.POST("/updates/", h.UpdateBatchMetrics("secret-key"))

	requestBody := []byte(`[{"id":"metric1","type":"gauge","value":10.5}]`)
	req, _ := http.NewRequest("POST", "/updates/", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output: 500
}

func Example_metricsListHandler() {
	router := gin.Default()

	ss := &config.Server{
		FlagAddress:     "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "tmp/metrics-db.json",
		FlagRestore:     true,
		DatabaseDSN:     "",
		SecretKey:       "secret-key",
	}

	s := filememory.NewMemStorage(true, ss)
	h := NewHandler(s)

	router.GET("/", h.MetricsListHandler())

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	fmt.Printf("Content Type: %v\n", resp.Header.Get("Content-Type"))
	// Output: Content Type: text/html; charset=utf-8
}

func Example_getMetricsJSONHandler() {
	router := gin.Default()

	ss := &config.Server{
		FlagAddress:     "localhost:8080",
		StoreInterval:   300,
		FileStoragePath: "tmp/metrics-db.json",
		FlagRestore:     true,
		DatabaseDSN:     "",
		SecretKey:       "your-secret-key",
	}
	s := filememory.NewMemStorage(true, ss)
	h := NewHandler(s)
	router.POST("/value/", h.GetMetricsJSONHandler("secret-key"))

	ts := httptest.NewServer(router)
	defer ts.Close()

	var d int64 = 4
	metricRequest := formatter.Metric{ID: "metric1", MType: "counter", Delta: &d, Value: nil}

	data, _ := json.Marshal(metricRequest)
	req, _ := http.NewRequest("POST", ts.URL+"/value/", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	fmt.Printf("Status Code: %v\n", resp.StatusCode)
	// Output: Status Code: 200
}
