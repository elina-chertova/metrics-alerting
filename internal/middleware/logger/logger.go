// Package logger provides logging functionalities for the application.
package logger

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Log is a global logger instance used throughout the application.
// It is initialized using the LogInit function.
var Log *zap.Logger = zap.NewNop()

// LogInit initializes the global logger with the specified logging level.
// It configures the logger based on the zap production configuration
// and sets the logging level according to the provided string.
func LogInit(level string) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		log.Printf("Error parsing logger level: %v", err)
		Log = nil
		return
	}
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()
	configuration := zap.NewProductionConfig()
	configuration.Level = lvl
	zl := zap.Must(configuration.Build())
	Log = zl
}

// Info logs an informational message with optional additional fields.
func Info(message string, fields ...zap.Field) {
	Log.Info(message, fields...)
}

// Error logs an error message with optional additional fields.
func Error(message string, fields ...zap.Field) {
	Log.Error(message, fields...)
}

// RequestLogger returns a Gin middleware function for logging HTTP requests.
// It logs the method, path, latency, size, and status of each request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		var size any

		switch c.Request.Method {
		case http.MethodPost:
			size = c.Request.ContentLength
		case http.MethodGet:
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
