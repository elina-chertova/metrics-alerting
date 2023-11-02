package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/security"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
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

	h := buildStorage(serverConfig, router)

	router.POST(
		"/updates/",
		security.HashCheckMiddleware(serverConfig.SecretKey),
		h.UpdateBatchMetrics(serverConfig.SecretKey),
	)
	router.POST(
		"/update/",
		security.HashCheckMiddleware(serverConfig.SecretKey),
		h.MetricsJSONHandler(serverConfig.SecretKey),
	)
	router.POST("/update/:metricType/:metricName/:metricValue", h.MetricsTextPlainHandler())
	router.GET(
		"/value/:metricType/:metricName",
		h.GetMetricsTextPlainHandler(serverConfig.SecretKey),
	)
	router.POST("/value/", h.GetMetricsJSONHandler(serverConfig.SecretKey))
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

func buildStorage(config *config.Server, router *gin.Engine) *handlers.Handler {
	if config.DatabaseDSN != "" {
		connection := db.Connect(config.DatabaseDSN)
		database := handlers.NewHandlerDB(connection)
		router.GET("/ping", database.PingDB())
		return handlers.NewHandler(connection)

	} else {
		s := filememory.NewMemStorage(true, config)
		return handlers.NewHandler(s)
	}
}
