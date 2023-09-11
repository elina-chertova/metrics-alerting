package handlers

import (
	"github.com/elina-chertova/metrics-alerting.git/cmd/storage"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"net/http"
	"strconv"
)

func MetricsListHandler(st *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		st.GaugeMu.Lock()
		st.CounterMu.Lock()

		tmpl, err := template.New("data").Parse("<!DOCTYPE html>\n<html>\n\n<head>\n    <title>Metric List</title>\n</head>\n\n<body>\n<ul>\n    {{ range $key, $value := .MetricsC }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n    {{ range $key, $value := .MetricsG }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n</ul>\n</body>\n\n</html>")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load template")
			return
		}

		err = tmpl.Execute(c.Writer, storage.MetricsData{
			MetricsC: st.Counter,
			MetricsG: st.Gauge,
		})
		st.CounterMu.Unlock()
		st.GaugeMu.Unlock()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}
	}
}

func GetMetricsHandler(st *storage.MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res any
		metricType := c.Param("metricType")
		metricName := c.Param("metricName")
		switch metricType {
		case storage.Gauge:
			_, ok := st.Gauge[metricName]
			if !ok {
				c.Status(http.StatusNotFound)
			}
			res = st.Gauge[metricName]
		case storage.Counter:
			_, ok := st.Gauge[metricName]
			if !ok {
				c.Status(http.StatusNotFound)
			}
			res = st.Counter[metricName]
		}
		resp, err := json.Marshal(res)
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
				st.GaugeMu.Lock()
				st.Gauge[metricName] = convertedMetricValueFloat
				st.GaugeMu.Unlock()
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case storage.Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				st.CounterMu.Lock()
				_, ok := st.Counter[metricName]
				if ok {
					st.Counter[metricName] += int64(convertedMetricValueInt)
				} else {
					st.Counter[metricName] = int64(convertedMetricValueInt)
				}
				st.CounterMu.Unlock()
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
