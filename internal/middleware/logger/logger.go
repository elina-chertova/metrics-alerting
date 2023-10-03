package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

var Log *zap.Logger

func LogInit(level string) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		_ = fmt.Errorf("error parsing logger level", err)
	}
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()
	configuration := zap.NewProductionConfig()
	configuration.Level = lvl
	zl := zap.Must(configuration.Build())
	Log = zl
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		var size any
		if c.Request.Method == http.MethodPost {
			size = c.Request.ContentLength
		} else if c.Request.Method == http.MethodGet {
			size = c.Writer.Size()
		}
		latency := time.Since(t)
		Log.Info(
			"got HTTP request info",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.String("latency", strconv.FormatInt(int64(latency), 10)),
			zap.Any("size", size),
			zap.Int("status", c.Writer.Status()),
		)
	}
}
