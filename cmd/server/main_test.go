package main

import (
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
	storage := MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(storage.MetricsHandler)
			h(w, request)
			result := w.Result()
			assert.Equal(t, result.StatusCode, tt.expected)
		})
	}
}

func Test_checkPost(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := checkPost(mockHandler)

	tests := []struct {
		name     string
		method   string
		expected int
	}{
		{
			name:     "Method Post",
			method:   http.MethodPost,
			expected: http.StatusOK,
		},
		{
			name:     "Method Get",
			method:   http.MethodGet,
			expected: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "http://localhost:8080/update/counter/metric2/10", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, request)

			assert.Equal(t, rr.Code, tt.expected)
		})
	}
}
