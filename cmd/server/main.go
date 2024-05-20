package main

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/handlers/rest"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/compression"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/security"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/subnet"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	if serverConfig.TrustedSubnet != "" {
		router.Use(subnet.TrustedIPMiddleware(serverConfig.TrustedSubnet))
	}

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
	srv := &http.Server{
		Addr:    serverConfig.FlagAddress,
		Handler: router,
	}

	//if err := router.Run(serverConfig.FlagAddress); err != nil {
	//	return err
	//}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-quit

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}

		log.Println("Server exiting")
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	return nil
}

func buildStorage(config *config.Server, router *gin.Engine) *rest.Handler {
	if config.DatabaseDSN != "" {
		connection := db.Connect(config.DatabaseDSN)
		database := rest.NewHandlerDB(connection)
		router.GET("/ping", database.PingDB())
		return rest.NewHandler(connection)

	} else {
		s := filememory.NewMemStorage(true, config)
		return rest.NewHandler(s)
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
