package compression

//
//type gzipWriter struct {
//	gin.ResponseWriter
//	Writer io.Writer
//}
//
//func (w gzipWriter) Write(b []byte) (int, error) {
//	return w.Writer.Write(b)
//}
//
//func GzipHandle() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		if !strings.Contains(
//			c.Request.Header.Get("Accept-Encoding"),
//			"gzip",
//		) && c.Request.Method != http.MethodGet {
//			c.Next()
//			return
//		}
//
//		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
//		if err != nil {
//			io.WriteString(c.Writer, err.Error())
//			return
//		}
//		defer gz.Close()
//
//		c.Header("Content-Encoding", "gzip")
//		c.Writer = gzipWriter{ResponseWriter: c.Writer, Writer: gz}
//		c.Next()
//	}
//}

//func LengthHandle(w http.ResponseWriter, r *http.Request) {
//	// создаём *gzip.Reader, который будет читать тело запроса
//	// и распаковывать его
//	gz, err := gzip.NewReader(r.Body)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	defer gz.Close()
//
//	// при чтении вернётся распакованный слайс байт
//	body, err := io.ReadAll(gz)
//	if err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//}
