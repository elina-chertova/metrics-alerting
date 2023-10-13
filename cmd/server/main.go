package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/fileMemory"
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
	logger.LogInit("info")
	router := gin.Default()

	router.Use(logger.RequestLogger())
	router.Use(compression.GzipHandle())
	var h *handlers.Handler

	if serverConfig.DatabaseDSN != "" {
		connection := db.Connect(serverConfig.DatabaseDSN)
		router.GET("/ping", connection.PingDB())
		h = handlers.NewHandler(connection)
	} else {
		s := fileMemory.NewMemStorage(true, serverConfig)
		h = handlers.NewHandler(s)
	}

	router.POST("/update/", h.MetricsJSONHandler())
	router.POST("/update/:metricType/:metricName/:metricValue", h.MetricsTextPlainHandler())
	router.GET("/value/:metricType/:metricName", h.GetMetricsTextPlainHandler())
	router.POST("/value/", h.GetMetricsJSONHandler())
	router.GET("/", h.MetricsListHandler())
	router.NoRoute(
		func(c *gin.Context) {
			c.String(http.StatusNotFound, "Page not found")
		},
	)

	if err := router.Run(serverConfig.FlagAddress); err != nil {
		return err
	}

	return nil
}
