package main

import (
	"github.com/elina-chertova/metrics-alerting.git/cmd/server/handlers"
	"github.com/elina-chertova/metrics-alerting.git/cmd/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	router := gin.Default()
	s := &storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(s))
	router.GET("/value/:metricType/:metricName", handlers.GetMetricsHandler(s))
	router.GET("/", handlers.MetricsListHandler(s))
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Page not found")
	})
	return router.Run("localhost:8080")
}
