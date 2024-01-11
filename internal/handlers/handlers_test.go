package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

//func TestGetMetricsJSONHandler(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	router := gin.New()
//
//	serverConfig := config.NewServer()
//	memStorage := filememory.NewMemStorage(true, serverConfig)
//
//	h := NewHandler(memStorage)
//	router.POST("/value/", h.GetMetricsJSONHandler("your-secret-key"))
//	// Additional test cases
//	tests := []struct {
//		name           string
//		requestBody    []byte // using raw JSON string for more control
//		expectedStatus int
//		expectedBody   string // you can expect specific error messages if you want
//	}{
//		{
//			name:           "Invalid Metric Type",
//			requestBody:    []byte(`{"ID": "metric2", "MType": "invalid", "Delta": 5}`),
//			expectedStatus: http.StatusBadRequest,
//			expectedBody:   "", // Replace with expected error message if necessary
//		},
//		{
//			name:           "Missing Metric ID",
//			requestBody:    []byte(`{"MType": "counter", "Delta": 5}`),
//			expectedStatus: http.StatusBadRequest,
//			expectedBody:   "", // Replace with expected error message if necessary
//		},
//		{
//			name:           "Invalid JSON Format",
//			requestBody:    []byte(`{"ID": "metric3", "MType": "counter", "Delta": "invalid"}`),
//			expectedStatus: http.StatusBadRequest,
//			expectedBody:   "", // Replace with expected error message if necessary
//		},
//		// Add more test cases as needed
//	}
//
//	// Run test cases
//	for _, tc := range tests {
//		t.Run(
//			tc.name, func(t *testing.T) {
//				// Create a request with JSON body
//				req := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(tc.requestBody))
//				req.Header.Set("Content-Type", "application/json")
//
//				// Record response
//				w := httptest.NewRecorder()
//				router.ServeHTTP(w, req)
//
//				// Validate response
//				assert.Equal(
//					t,
//					tc.expectedStatus,
//					w.Code,
//					"Expected and actual status codes should match",
//				)
//
//				if tc.expectedBody != "" {
//					responseBody, _ := ioutil.ReadAll(w.Body)
//					assert.Contains(
//						t,
//						string(responseBody),
//						tc.expectedBody,
//						"Response body should contain the expected message",
//					)
//				}
//				// Add more assertions as needed
//			},
//		)
//	}
//
//	//tests := []struct {
//	//	name           string
//	//	metric         f.Metric
//	//	expectedStatus int
//	//}{
//	//	{
//	//		name: "Valid Counter Metric",
//	//		metric: f.Metric{
//	//			ID:    "metric1",
//	//			MType: config.Counter,
//	//			Delta: new(int64),
//	//		},
//	//		expectedStatus: http.StatusOK,
//	//	},
//	//}
//	//
//	//for _, tc := range tests {
//	//	t.Run(
//	//		tc.name, func(t *testing.T) {
//	//			jsonData, _ := json.Marshal(tc.metric)
//	//
//	//			req := httptest.NewRequest("POST", "/value/", bytes.NewBuffer(jsonData))
//	//			req.Header.Set("Content-Type", "application/json")
//	//
//	//			w := httptest.NewRecorder()
//	//			router.ServeHTTP(w, req)
//	//
//	//			assert.Equal(
//	//				t,
//	//				tc.expectedStatus,
//	//				w.Code,
//	//				"Expected and actual status codes should match",
//	//			)
//	//		},
//	//	)
//	//}
//}
