package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			storage := &MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			router.POST("/update/:metricType/:metricName/:metricValue", MetricsHandler(storage))

			request := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)
			result := w.Result()
			assert.Equal(t, tt.expected, result.StatusCode)
		})
	}
}