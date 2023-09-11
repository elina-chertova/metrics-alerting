package main

import (
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"html/template"
	"net/http"
	"strconv"
	"sync"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type MetricsData struct {
	MetricsC map[string]int64
	MetricsG map[string]float64
}

type MemStorage struct {
	gaugeMu   sync.Mutex
	Gauge     map[string]float64 `json:"gauge"`
	counterMu sync.Mutex
	Counter   map[string]int64 `json:"counter"`
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	router := gin.Default()
	s := &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	router.POST("/update/:metricType/:metricName/:metricValue", MetricsHandler(s))
	router.GET("/value/:metricType/:metricName", GetMetricsHandler(s))
	router.GET("/", MetricsListHandler(s))
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Page not found")
	})
	return router.Run("localhost:8080")
}

func MetricsListHandler(storage *MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		storage.gaugeMu.Lock()
		storage.counterMu.Lock()

		tmpl, err := template.New("data").Parse("<!DOCTYPE html>\n<html>\n\n<head>\n    <title>Metric List</title>\n</head>\n\n<body>\n<ul>\n    {{ range $key, $value := .MetricsC }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n    {{ range $key, $value := .MetricsG }}\n    <p>{{$key}}: {{$value}}</p>\n    {{ end }}\n</ul>\n</body>\n\n</html>")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load template")
			return
		}

		err = tmpl.Execute(c.Writer, MetricsData{
			MetricsC: storage.Counter,
			MetricsG: storage.Gauge,
		})
		storage.counterMu.Unlock()
		storage.gaugeMu.Unlock()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to render template")
			return
		}
	}
}

func GetMetricsHandler(storage *MemStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var res any
		metricType := c.Param("metricType")
		metricName := c.Param("metricName")
		switch metricType {
		case Gauge:
			_, ok := storage.Gauge[metricName]
			if !ok {
				c.Status(http.StatusNotFound)
			}
			res = storage.Gauge[metricName]
		case Counter:
			_, ok := storage.Gauge[metricName]
			if !ok {
				c.Status(http.StatusNotFound)
			}
			res = storage.Counter[metricName]
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

func MetricsHandler(storage *MemStorage) gin.HandlerFunc {
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
		case Gauge:
			if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
				storage.gaugeMu.Lock()
				storage.Gauge[metricName] = convertedMetricValueFloat
				storage.gaugeMu.Unlock()
			} else {
				c.Status(http.StatusBadRequest)
				return
			}
		case Counter:
			if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
				storage.counterMu.Lock()
				_, ok := storage.Counter[metricName]
				if ok {
					storage.Counter[metricName] += int64(convertedMetricValueInt)
				} else {
					storage.Counter[metricName] = int64(convertedMetricValueInt)
				}
				storage.counterMu.Unlock()
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
