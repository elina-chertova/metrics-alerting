package subnet

import (
	"errors"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
)

var ErrUntrustedSubnet = errors.New("untrusted subnet")

func TrustedIPMiddleware(trustedIP string) gin.HandlerFunc {
	_, trustedIPNet, err := net.ParseCIDR(trustedIP)
	if err != nil {
		logger.Log.Fatal("Invalid trusted subnet", zap.Error(err))
	}
	return func(c *gin.Context) {
		requestIP := c.Request.Header.Get("X-Real-IP")
		if trustedIP == "" {
			c.Next()
			return
		}
		if requestIP == "" {
			logger.Log.Info("Missing X-Real-IP header")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrUntrustedSubnet.Error()})
			return
		}

		ip := net.ParseIP(requestIP)
		if ip == nil {
			logger.Log.Info(
				"Invalid IP address in X-Real-IP header",
				zap.String("requestIP", requestIP),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrUntrustedSubnet.Error()})
			return
		}

		if !trustedIPNet.Contains(ip) {
			logger.Log.Info(
				ErrUntrustedSubnet.Error(),
				zap.String("requestIP", requestIP),
				zap.String("trustedSubnet", trustedIP),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrUntrustedSubnet.Error()})
			return
		}
		c.Next()
	}
}
