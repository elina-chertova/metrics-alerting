package handlers

import (
	"errors"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

var (
	ErrInvalidJSON        = errors.New("invalid JSON data")
	ErrUnsupportedMetric  = errors.New("unsupported metric type")
	ErrDeltaNil           = errors.New("delta is nil, skipping update")
	ErrValueNil           = errors.New("value is nil, skipping update")
	ErrFailedJSONCreating = errors.New("failed JSON creation")
)

type metricsStorage interface {
	UpdateCounter(name string, value int64, ok bool) error
	UpdateGauge(name string, value float64) error
	GetCounter(name string) (int64, bool, error)
	GetGauge(name string) (float64, bool, error)
	GetMetrics() (map[string]int64, map[string]float64)
	InsertBatchMetrics([]f.Metric) error
}

type Handler struct {
	memStorage metricsStorage
}

func NewHandler(st metricsStorage) *Handler {
	return &Handler{st}
}

func (h *Handler) UpdateBatchMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m []f.Metric
		var reader io.Reader = c.Request.Body
		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(c.Writer, "error reading request body", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &m); err != nil {
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}

		err = h.memStorage.InsertBatchMetrics(m)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed data insert")
			return
		}

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")

		responseJSON, err := json.Marshal(make(map[string]interface{}))
		if err != nil {
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}

		c.Writer.Write(responseJSON)
	}
}

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

func (h *Handler) GetMetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m f.Metric
		var err error
		if err := c.ShouldBindJSON(&m); err != nil {
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
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Gauge:
			val2, _, err = h.memStorage.GetGauge(m.ID)
			metric = f.Metric{ID: m.ID, MType: config.Gauge, Value: &val2}
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		out, err := json.Marshal(metric)
		if err != nil {
			c.String(http.StatusInternalServerError, ErrFailedJSONCreating.Error())
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

func (h *Handler) GetMetricsTextPlainHandler() gin.HandlerFunc {
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
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _, err = h.memStorage.GetGauge(metricName)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Counter:
			_, ok, err = h.memStorage.GetCounter(metricName)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _, err = h.memStorage.GetCounter(metricName)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		resp, err := json.Marshal(value)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(resp)
	}
}

func (h *Handler) MetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			m   f.Metric
			err error
			ok  bool
		)
		var reader io.Reader = c.Request.Body
		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(c.Writer, "error reading request body", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &m); err != nil {
			http.Error(c.Writer, ErrInvalidJSON.Error(), http.StatusBadRequest)
			return
		}

		var returnedMetric f.Metric

		switch m.MType {
		case config.Counter:
			if m.Delta == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": ErrDeltaNil.Error()})
				return
			}
			_, ok, err = h.memStorage.GetCounter(m.ID)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			var v1 = *m.Delta
			err = h.memStorage.UpdateCounter(m.ID, v1, ok)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			v1, _, err = h.memStorage.GetCounter(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: config.Counter, Delta: &v1}
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		case config.Gauge:
			if m.Value == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": ErrValueNil.Error()})
				return
			}
			var v2 = *m.Value
			err = h.memStorage.UpdateGauge(m.ID, v2)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			v2, _, err = h.memStorage.GetGauge(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: config.Gauge, Value: &v2}
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": ErrUnsupportedMetric.Error()})
			return
		}

		out, err := json.Marshal(returnedMetric)
		if err != nil {
			c.String(http.StatusInternalServerError, ErrFailedJSONCreating.Error())
		}

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

func (h *Handler) MetricsTextPlainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			err error
			ok  bool
		)
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
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case config.Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				_, ok, err = h.memStorage.GetCounter(metricName)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				err = h.memStorage.UpdateCounter(metricName, int64(convertedMetricValueInt), ok)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		default:
			c.Status(http.StatusBadRequest)
			return
		}
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.Header("Content-Type", "text/plain")

		c.Status(http.StatusOK)
	}
}

type Database struct {
	db *db.DB
}

func NewDatabase(d *db.DB) *Database {
	return &Database{db: d}
}

func (db *Database) PingDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.db.Database.DB()
		if err != nil {
			handleDBError(c, "failed to get database connection", err)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			handleDBError(c, "failed to ping the database", err)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{"message": "Successfully connected to the database and pinged it"},
		)
	}
}

func handleDBError(c *gin.Context, message string, err error) {
	log.Printf("%s: %v", message, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
