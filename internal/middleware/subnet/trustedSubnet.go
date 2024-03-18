package subnet

import (
	"errors"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

var ErrUntrustedSubnet = errors.New("untrusted subnet")

func TrustedIPMiddleware(trustedIP string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIP := c.Request.Header.Get("X-Real-IP")
		if trustedIP == "" {
			c.Next()
			return
		}

		if requestIP != trustedIP {
			logger.Log.Info(
				ErrUntrustedSubnet.Error(),
				zap.String("requestIP", requestIP),
				zap.String("trustedIP", trustedIP),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrUntrustedSubnet.Error()})
			return
		}
		c.Next()
	}
}
