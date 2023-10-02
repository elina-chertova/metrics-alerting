package compression

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(
			c.Request.Header.Get("Accept-Encoding"),
			"gzip",
		) {
			c.Next()
			return
		}

		if c.Request.Header.Get("Content-Type") != "text/plain" && c.Request.Header.Get("Content-Type") != "application/json" && c.Request.Header.Get("Content-Type") != "html/text" {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			io.WriteString(c.Writer, err.Error())
			return
		}
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}
		c.Next()
	}
}
