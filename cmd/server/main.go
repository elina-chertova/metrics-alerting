package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	serverConfig := config.NewServer()
	logger.LogInit()
	router := gin.Default()
	router.Use(logger.RequestLogger())
	router.Use(compression.GzipHandle())
	s := storage.NewMemStorage()
	h := handlers.NewHandler(s)
	router.POST("/update", h.MetricsJSONHandler())
	router.POST("/update/:metricType/:metricName/:metricValue", h.MetricsTextPlainHandler())
	router.GET("/value/:metricType/:metricName", h.GetMetricsTextPlainHandler())
	router.GET("/value/", h.GetMetricsJSONHandler())
	router.GET("/", h.MetricsListHandler())
	router.NoRoute(
		func(c *gin.Context) {
			c.String(http.StatusNotFound, "Page not found")
		},
	)
	return router.Run(serverConfig.FlagAddress)
}
