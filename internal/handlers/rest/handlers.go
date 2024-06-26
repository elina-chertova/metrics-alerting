// Package rest Package handlers provides HTTP handlers for various endpoints
// in a metrics alerting application.
package rest

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/elina-chertova/metrics-alerting.git/internal/asymencrypt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	serviceInterface "github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/security"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

var (
	ErrInvalidJSON        = errors.New("invalid JSON data")
	ErrUnsupportedMetric  = errors.New("unsupported metric type")
	ErrDeltaNil           = errors.New("delta is nil, skipping update")
	ErrValueNil           = errors.New("value is nil, skipping update")
	ErrFailedJSONCreating = errors.New("failed JSON creation")
	ErrReadReqBody        = errors.New("error reading request body")
)

// database defines an interface for interacting with a database.
type database interface {
	PingDB() gin.HandlerFunc
}

// Handler encapsulates handling logic for metric-related HTTP endpoints.
type Handler struct {
	memStorage serviceInterface.MetricsStorage
}

type HandlerDB struct {
	db database
}

// NewHandler creates a new Handler with the given metrics storage.
func NewHandler(st serviceInterface.MetricsStorage) *Handler {
	return &Handler{st}
}

// NewHandlerDB creates a new HandlerDB with the given database interface.
func NewHandlerDB(d database) *HandlerDB {
	return &HandlerDB{db: d}
}

func (db *HandlerDB) PingDB() gin.HandlerFunc {
	return db.db.PingDB()
}

// UpdateBatchMetrics creates a gin.HandlerFunc that handles batch updates
// of metric data. It processes JSON requests containing multiple metrics
// and updates them in the storage.
func (h *Handler) UpdateBatchMetrics(secretKey string, privateKeyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var m []f.Metric

		var reader io.Reader = c.Request.Body
		body, err := io.ReadAll(reader)
		if err != nil {
			logger.Error(ErrReadReqBody.Error(), zap.String("method", c.Request.Method))
			http.Error(c.Writer, ErrReadReqBody.Error(), http.StatusInternalServerError)
			return
		}

		if privateKeyPath != "" {
			decryptedBody, err := asymencrypt.DecryptDataWithPrivateKey(body, privateKeyPath)
			if err != nil {
				logger.Error("Failed to decrypt data", zap.String("method", c.Request.Method))
				http.Error(c.Writer, "Failed to decrypt data", http.StatusInternalServerError)
				return
			}
			body = decryptedBody
		}

		if err = json.Unmarshal(body, &m); err != nil {
			logger.Error(ErrInvalidJSON.Error(), zap.String("method", c.Request.Method))
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}

		err = h.memStorage.InsertBatchMetrics(m)
		if err != nil {
			logger.Error(err.Error(), zap.String("method", c.Request.Method))
			c.String(http.StatusInternalServerError, "Failed data insert")
			return
		}

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")

		responseJSON, err := json.Marshal(make(map[string]interface{}))
		if err != nil {
			logger.Error(ErrInvalidJSON.Error(), zap.String("method", c.Request.Method))
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}
		if secretKey != "" {
			c.Writer.Header().Set("HashSHA256", "")
		}

		c.Writer.Write(responseJSON)
	}
}

// MetricsListHandler creates a gin.HandlerFunc that serves a webpage displaying
// a list of all stored metrics. It renders the metrics data in an HTML template.
func (h *Handler) MetricsListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tmpl, err := template.New("data").Parse("<!DOCTYPE html>\n<html>\n\n<head>\n    <title>Metric List</title>\n</head>\n\n<body>\n<ul>\n    {{ range $key, $value := .MetricsC }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n    {{ range $key, $value := .MetricsG }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n</ul>\n</body>\n\n</html>")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load template")
			return
		}
		counter, gauge := h.memStorage.GetMetrics()
		err = tmpl.Execute(
			c.Writer, struct {
				MetricsC map[string]int64
				MetricsG map[string]float64
			}{
				MetricsC: counter,
				MetricsG: gauge,
			},
		)

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}
		c.Writer.Header().Set("Content-Type", "text/html")
	}
}

// GetMetricsJSONHandler creates a gin.HandlerFunc for retrieving a specific metric
// in JSON format. The handler reads a metric ID and type from the request
// and returns it as JSON.
func (h *Handler) GetMetricsJSONHandler(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var m f.Metric
		var err error
		if err = c.ShouldBindJSON(&m); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var metric f.Metric
		var val1 int64
		var val2 float64

		switch m.MType {
		case config.Counter:
			val1, _, err = h.memStorage.GetCounter(m.ID)
			metric = f.Metric{ID: m.ID, MType: config.Counter, Delta: &val1}
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Gauge:
			val2, _, err = h.memStorage.GetGauge(m.ID)
			metric = f.Metric{ID: m.ID, MType: config.Gauge, Value: &val2}
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			logger.Error(ErrUnsupportedMetric.Error(), zap.String("method", c.Request.Method))
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		out, err := json.Marshal(metric)
		if err != nil {
			c.String(http.StatusInternalServerError, ErrFailedJSONCreating.Error())
			return
		}
		if secretKey != "" {
			correctHash := security.Hash(string(out), []byte(secretKey))
			c.Writer.Header().Set("HashSHA256", correctHash)
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

// GetMetricsTextPlainHandler creates a gin.HandlerFunc for retrieving and updating
// a specific metric in plain text format. The handler reads metric details from
// the request URL, performs necessary operations (like updating or retrieving),
// and responds with the metric value in plain text.
func (h *Handler) GetMetricsTextPlainHandler(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			value any
			err   error
			ok    bool
		)
		metricType := c.Param("metricType")
		metricName := c.Param("metricName")

		switch metricType {
		case config.Gauge:
			_, ok, err = h.memStorage.GetGauge(metricName)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _, err = h.memStorage.GetGauge(metricName)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Counter:
			_, ok, err = h.memStorage.GetCounter(metricName)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _, err = h.memStorage.GetCounter(metricName)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			logger.Error(ErrUnsupportedMetric.Error(), zap.String("method", c.Request.Method))
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		resp, err := json.Marshal(value)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if secretKey != "" {
			correctHash := security.Hash(string(resp), []byte(secretKey))
			c.Writer.Header().Set("HashSHA256", correctHash)
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(resp)
	}
}

// MetricsJSONHandler creates a gin.HandlerFunc for processing incoming metric data
// in JSON format. The handler reads JSON formatted metric data from the request body,
// updates or retrieves the metric in storage, and responds with the updated metric data.
func (h *Handler) MetricsJSONHandler(secretKey string, privateKeyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			m   f.Metric
			err error
			ok  bool
		)

		var reader io.Reader = c.Request.Body
		body, err := io.ReadAll(reader)
		if err != nil {
			logger.Error("error reading request body", zap.String("method", c.Request.Method))
			http.Error(c.Writer, "error reading request body", http.StatusInternalServerError)
			return
		}

		if privateKeyPath != "" {
			decryptedBody, err := asymencrypt.DecryptDataWithPrivateKey(body, privateKeyPath)
			if err != nil {
				logger.Error("Failed to decrypt data", zap.String("method", c.Request.Method))
				http.Error(c.Writer, "Failed to decrypt data", http.StatusInternalServerError)
				return
			}
			body = decryptedBody
		}

		if err = json.Unmarshal(body, &m); err != nil {
			logger.Error(ErrInvalidJSON.Error(), zap.String("method", c.Request.Method))
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}

		var returnedMetric f.Metric

		switch m.MType {
		case config.Counter:
			if m.Delta == nil {
				logger.Error(ErrDeltaNil.Error(), zap.String("method", c.Request.Method))
				c.JSON(http.StatusBadRequest, gin.H{"error": ErrDeltaNil.Error()})
				return
			}
			_, ok, err = h.memStorage.GetCounter(m.ID)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			var v1 = *m.Delta
			err = h.memStorage.UpdateCounter(m.ID, v1, ok)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			v1, _, err = h.memStorage.GetCounter(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: config.Counter, Delta: &v1}
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Gauge:
			if m.Value == nil {
				logger.Error(ErrValueNil.Error(), zap.String("method", c.Request.Method))
				c.JSON(http.StatusBadRequest, gin.H{"error": ErrValueNil.Error()})
				return
			}
			var v2 = *m.Value
			err = h.memStorage.UpdateGauge(m.ID, v2)
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			v2, _, err = h.memStorage.GetGauge(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: config.Gauge, Value: &v2}
			if err != nil {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			logger.Error(ErrUnsupportedMetric.Error(), zap.String("method", c.Request.Method))
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		out, err := json.Marshal(returnedMetric)
		if err != nil {
			logger.Error(ErrFailedJSONCreating.Error(), zap.String("method", c.Request.Method))
			c.String(http.StatusInternalServerError, ErrFailedJSONCreating.Error())
		}
		if secretKey != "" {
			correctHash := security.Hash(string(out), []byte(secretKey))
			c.Writer.Header().Set("HashSHA256", correctHash)
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

// MetricsTextPlainHandler creates a gin.HandlerFunc that handles metric data
// submitted in plain text format. The handler parses the metric type, name,
// and value from the request, updates or retrieves the metric in storage,
// and sends back a plain text response.
func (h *Handler) MetricsTextPlainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ok bool
		if err := c.Request.ParseForm(); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		metricType := c.Param("metricType")
		metricName := c.Param("metricName")
		metricValue := c.Param("metricValue")
		if metricName == "" {
			c.Status(http.StatusNotFound)
			return
		}
		switch metricType {
		case config.Gauge:
			if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
				err = h.memStorage.UpdateGauge(metricName, convertedMetricValueFloat)
				if err != nil {
					logger.Error(err.Error(), zap.String("method", c.Request.Method))
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
			} else {
				logger.Error(err.Error(), zap.String("method", c.Request.Method))
				c.Status(http.StatusBadRequest)
				return
			}
		case config.Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				_, ok, err = h.memStorage.GetCounter(metricName)
				if err != nil {
					logger.Error(err.Error(), zap.String("method", c.Request.Method))
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				err = h.memStorage.UpdateCounter(metricName, int64(convertedMetricValueInt), ok)
				if err != nil {
					logger.Error(err.Error(), zap.String("method", c.Request.Method))
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
			} else {
				logger.Error(
					err.Error(),
					zap.String("method", c.Request.Method),
					zap.Int("status", http.StatusBadRequest),
				)
				c.Status(http.StatusBadRequest)
				return
			}
		default:
			logger.Error(
				ErrUnsupportedMetric.Error(),
				zap.String("method", c.Request.Method),
				zap.Int("status", http.StatusBadRequest),
			)
			c.Status(http.StatusBadRequest)

			return
		}

		c.Header("Content-Type", "text/plain")
		c.Status(http.StatusOK)
	}
}
