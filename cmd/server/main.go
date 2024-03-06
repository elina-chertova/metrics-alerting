package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/security"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"github.com/gin-gonic/gin"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:%s\n", buildVersion)
	fmt.Printf("Build date:%s\n", buildDate)
	fmt.Printf("Build commit:%s\n", buildCommit)

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

	RegisterPprofRoutes(router)
	router.POST(
		"/updates/",
		security.HashCheckMiddleware(serverConfig.SecretKey),
		h.UpdateBatchMetrics(serverConfig.SecretKey, serverConfig.CryptoKey),
	)
	router.POST(
		"/update/",
		security.HashCheckMiddleware(serverConfig.SecretKey),
		h.MetricsJSONHandler(serverConfig.SecretKey, serverConfig.CryptoKey),
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

func RegisterPprofRoutes(router *gin.Engine) {
	router.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	router.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	router.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	router.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	router.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))

	router.GET("/debug/pprof/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
	router.GET("/debug/pprof/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
	router.GET("/debug/pprof/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
	router.GET("/debug/pprof/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
}
