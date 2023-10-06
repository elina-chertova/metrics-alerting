package handlers

import (
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/metrics"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

type metricsStorage interface {
	UpdateCounter(name string, value int64, ok bool)
	UpdateGauge(name string, value float64)
	GetCounter(name string) (int64, bool)
	GetGauge(name string) (float64, bool)
}

type handler struct {
	memStorage *metrics.MemStorage
}

func NewHandler(st *metrics.MemStorage) *handler {
	return &handler{st}
}

func (h *handler) MetricsListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.memStorage.LockGauge()
		h.memStorage.LockCounter()
		defer h.memStorage.UnlockGauge()
		defer h.memStorage.UnlockCounter()

		tmpl, err := template.New("data").Parse("<!DOCTYPE html>\n<html>\n\n<head>\n    <title>Metric List</title>\n</head>\n\n<body>\n<ul>\n    {{ range $key, $value := .MetricsC }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n    {{ range $key, $value := .MetricsG }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n</ul>\n</body>\n\n</html>")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load template")
			return
		}

		err = tmpl.Execute(
			c.Writer, metrics.MetricsData{
				MetricsC: h.memStorage.Counter,
				MetricsG: h.memStorage.Gauge,
			},
		)

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}

		c.Writer.Header().Set("Content-Type", "text/html")
	}
}

func (h *handler) GetMetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m f.Metric
		if err := c.ShouldBindJSON(&m); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var metric f.Metric
		var val1 int64
		var val2 float64

		switch m.MType {
		case metrics.Counter:
			val1, _ = h.memStorage.GetCounter(m.ID)
			metric = f.Metric{ID: m.ID, MType: metrics.Counter, Delta: &val1}
		case metrics.Gauge:
			val2, _ = h.memStorage.GetGauge(m.ID)
			metric = f.Metric{ID: m.ID, MType: metrics.Gauge, Value: &val2}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported metric type"})
			return
		}
		out, err := json.Marshal(metric)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed json creating")
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

func (h *handler) GetMetricsTextPlainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var value any
		metricType := c.Param("metricType")
		metricName := c.Param("metricName")

		switch metricType {
		case metrics.Gauge:
			_, ok := h.memStorage.GetGauge(metricName)
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _ = h.memStorage.GetGauge(metricName)

		case metrics.Counter:
			_, ok := h.memStorage.GetCounter(metricName)
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _ = h.memStorage.GetCounter(metricName)
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

func (h *handler) MetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m f.Metric
		var reader io.Reader = c.Request.Body
		body, err := io.ReadAll(reader)
		if err != nil {
			http.Error(c.Writer, "error reading request body", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(body, &m); err != nil {
			http.Error(c.Writer, "invalid JSON data", http.StatusBadRequest)
			return
		}

		var returnedMetric f.Metric

		switch m.MType {
		case metrics.Counter:
			if m.Delta == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Delta is nil, skipping update"})
				return
			}
			_, ok := h.memStorage.GetCounter(m.ID)
			var v1 = *m.Delta
			h.memStorage.UpdateCounter(m.ID, v1, ok)

			v1, _ = h.memStorage.GetCounter(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: metrics.Counter, Delta: &v1}
		case metrics.Gauge:
			if m.Value == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Value is nil, skipping update"})
				return
			}
			var v2 = *m.Value
			h.memStorage.UpdateGauge(m.ID, v2)

			v2, _ = h.memStorage.GetGauge(m.ID)
			returnedMetric = f.Metric{ID: m.ID, MType: metrics.Gauge, Value: &v2}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported metric type"})
			return
		}
		out, err := json.Marshal(returnedMetric)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed json creating")
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(out)
	}
}

func (h *handler) MetricsTextPlainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

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
		case metrics.Gauge:
			if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
				h.memStorage.UpdateGauge(metricName, convertedMetricValueFloat)
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case metrics.Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				_, ok := h.memStorage.GetCounter(metricName)
				h.memStorage.UpdateCounter(metricName, int64(convertedMetricValueInt), ok)
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		default:
			c.Status(http.StatusBadRequest)
			return
		}
		c.Header("Content-Type", "text/plain")

		c.Status(http.StatusOK)
	}
}
