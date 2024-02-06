// Package compression provides middleware for handling gzip compression in HTTP requests and responses.
package compression

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	gzip "github.com/klauspost/pgzip"
)

// gzipWriter wraps gin.ResponseWriter and a generic io.Writer to enable gzip
// compression on HTTP responses.
type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

// Write compresses the given byte slice and writes it to the HTTP response.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipHandle returns a Gin middleware function for handling gzip compression.
// This middleware also handles decompressing gzip-encoded request bodies.
func GzipHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.Request.Header.Get("Content-Type")
		if contentType == "text/html" || c.Request.RequestURI == "/" {
			c.Writer.Header().Set("Content-Type", "text/html")
		} else if contentType == "text/plain" {
			c.Writer.Header().Set("Content-Type", "text/plain")
		} else if contentType == "application/json" {
			c.Writer.Header().Set("Content-Type", "application/json")
		} else {
			c.Next()
			return
		}

		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		var gz *gzip.Writer
		if supportsGzip {
			var err error
			gz, err = gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				io.WriteString(c.Writer, err.Error())
				return
			}
			defer gz.Close()
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			body, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			c.Request.Body = body
			defer body.Close()
		}

		if gz != nil {
			c.Header("Content-Encoding", "gzip")
			c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}
		}

		c.Next()
	}
}
