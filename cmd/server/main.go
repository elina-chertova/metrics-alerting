package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/metrics"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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

	s := metrics.NewMemStorage()
	h := handlers.NewHandler(s)
	st := store.NewStorager(s)
	if serverConfig.FlagRestore {
		st.Load(serverConfig.FileStoragePath)
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

	go func() {
		for {
			time.Sleep(time.Duration(serverConfig.StoreInterval) * time.Second)
			st.Save(serverConfig.FileStoragePath)
		}
	}()

	if err := router.Run(serverConfig.FlagAddress); err != nil {
		return err
	}

	return nil
}
