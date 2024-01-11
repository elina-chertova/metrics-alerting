package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

//
//func TestGetMetricsTextPlainHandler(t *testing.T) {
//	tests := []struct {
//		name     string
//		path     string
//		expected int
//	}{
//		{
//			name:     "Valid Gauge Metric",
//			path:     "/value/gauge/metric1",
//			expected: http.StatusOK,
//		},
//		{
//			name:     "Valid Counter Metric",
//			path:     "/value/counter/metric2",
//			expected: http.StatusOK,
//		},
//		{
//			name:     "Invalid Metric Type",
//			path:     "/value/invalid/metric3",
//			expected: http.StatusBadRequest,
//		},
//		{
//			name:     "Metric Not Found",
//			path:     "/value/gauge/nonexistent",
//			expected: http.StatusNotFound,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(
//			tt.name, func(t *testing.T) {
//				router := gin.Default()
//				st := filememory.NewMemStorage(false, nil)
//				h := NewHandler(st)
//				router.GET("/value/:metricType/:metricName", h.GetMetricsTextPlainHandler(""))
//
//				request := httptest.NewRequest(http.MethodGet, tt.path, nil)
//				w := httptest.NewRecorder()
//
//				router.ServeHTTP(w, request)
//				result := w.Result()
//				defer result.Body.Close()
//				assert.Equal(t, tt.expected, result.StatusCode)
//			},
//		)
//	}
//}

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
