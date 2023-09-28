package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/logger"
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
	s := storage.NewMemStorage()
	router.POST("/update/:metricType/:metricName/:metricValue", handlers.MetricsHandler(s))
	router.GET("/value/:metricType/:metricName", handlers.GetMetricsHandler(s))
	router.GET("/", handlers.MetricsListHandler(s))
	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Page not found")
	})
	return router.Run(serverConfig.FlagAddress)
}
