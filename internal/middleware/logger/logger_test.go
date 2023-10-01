package logger

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestRequestLogger(t *testing.T) {
	LogInit()

	r := gin.Default()
	r.Use(RequestLogger())

	r.Run(":8080")
}
