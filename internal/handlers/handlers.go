package handlers

import (
	"compress/gzip"
	"fmt"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type metricsStorage interface {
	UpdateCounter(name string, value int64, ok bool)
	UpdateGauge(name string, value float64)
	GetCounter(name string) (int64, bool)
	GetGauge(name string) (float64, bool)
}

type handler struct {
	memStorage *storage.MemStorage
}

func NewHandler(st *storage.MemStorage) *handler {
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
			c.Writer, storage.MetricsData{
				MetricsC: h.memStorage.Counter,
				MetricsG: h.memStorage.Gauge,
			},
		)

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}
	}
}

func (h *handler) GetMetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var metrics []f.Metric
		for name, value := range h.memStorage.Gauge {
			val := value
			metrics = append(
				metrics, f.Metric{
					ID:    name,
					MType: storage.Gauge,
					Value: &val,
				},
			)
		}
		for name, value := range h.memStorage.Counter {
			val := value
			metrics = append(
				metrics, f.Metric{
					ID:    name,
					MType: storage.Counter,
					Delta: &val,
				},
			)
		}
		out, err := json.Marshal(metrics)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
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
		case storage.Gauge:
			_, ok := h.memStorage.GetGauge(metricName)
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _ = h.memStorage.GetGauge(metricName)

		case storage.Counter:
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

type ResMetric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta,omitempty"`
	Value float64 `json:"value,omitempty"`
}

func (h *handler) MetricsJSONHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m []ResMetric
		if err := c.Request.ParseForm(); err != nil {
			fmt.Println("here1")
			c.Status(http.StatusBadRequest)
			return
		}

		if strings.Contains(
			c.Request.Header.Get("Content-Encoding"),
			"gzip",
		) {
			fmt.Println("here2")
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := json.NewDecoder(gz).Decode(&m); err != nil {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if err := c.ShouldBindJSON(&m); err != nil {
				fmt.Println("here4")
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			_ = json.NewDecoder(c.Request.Body).Decode(&m)
			c.Writer.Header().Set("Content-Type", "application/json")
		}

		for _, metric := range m {
			fmt.Println("here5")
			switch metric.MType {
			case storage.Counter:
				_, ok := h.memStorage.GetCounter(metric.ID)
				h.memStorage.UpdateCounter(metric.ID, metric.Delta, ok)
			case storage.Gauge:
				h.memStorage.UpdateGauge(metric.ID, metric.Value)
			default:
				c.Status(http.StatusBadRequest)
				return
			}

		}
		c.Header("Content-Type", "application/json")

		c.Status(http.StatusOK)
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
		case storage.Gauge:
			if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
				h.memStorage.UpdateGauge(metricName, convertedMetricValueFloat)
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case storage.Counter:
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
