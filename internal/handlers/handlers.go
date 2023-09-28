package handlers

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"net/http"
	"strconv"
)

type metricsStorage interface {
	UpdateCounter(name string, value int64, ok bool)
	UpdateGauge(name string, value float64)
	GetCounter(name string) (int64, bool)
	GetGauge(name string) (float64, bool)
}

func MetricsListHandler(st *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		st.LockGauge()
		st.LockCounter()
		defer st.UnlockGauge()
		defer st.UnlockCounter()

		tmpl, err := template.New("data").Parse("<!DOCTYPE html>\n<html>\n\n<head>\n    <title>Metric List</title>\n</head>\n\n<body>\n<ul>\n    {{ range $key, $value := .MetricsC }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n    {{ range $key, $value := .MetricsG }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n</ul>\n</body>\n\n</html>")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load template")
			return
		}

		err = tmpl.Execute(c.Writer, storage.MetricsData{
			MetricsC: st.Counter,
			MetricsG: st.Gauge,
		})

		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}
	}
}

func GetMetricsHandler(st *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var value any
		metricType := c.Param("metricType")
		metricName := c.Param("metricName")

		switch metricType {
		case storage.Gauge:
			_, ok := st.GetGauge(metricName)
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _ = st.GetGauge(metricName)

		case storage.Counter:
			_, ok := st.GetCounter(metricName)
			if !ok {
				c.Status(http.StatusNotFound)
				return
			}
			value, _ = st.GetCounter(metricName)
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

func MetricsHandler(st *storage.MemStorage) gin.HandlerFunc {
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
				st.UpdateGauge(metricName, convertedMetricValueFloat)
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case storage.Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				_, ok := st.GetCounter(metricName)
				st.UpdateCounter(metricName, int64(convertedMetricValueInt), ok)
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
