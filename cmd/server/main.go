package main

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	MethodPost = "POST"

	Gauge   = "gauge"
	Counter = "counter"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

type Middleware func(http.Handler) http.Handler

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func run() error {
	mux := http.NewServeMux()
	s := &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
	mux.Handle("/update/", Conveyor(http.HandlerFunc(s.MetricsHandler), checkPost))
	return http.ListenAndServe(`:8080`, mux)
}

func checkPost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

//func checkContentType(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		contentType := r.Header.Get("Content-Type")
//		if contentType != "text/plain" {
//			w.WriteHeader(http.StatusUnsupportedMediaType)
//			return
//		}
//		next.ServeHTTP(w, r)
//	})
//}

func (storage *MemStorage) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		//w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
	}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		w.WriteHeader(http.StatusNotFound)
		//http.Error(w, "some field is empty", http.StatusBadRequest)
		return
	}
	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	if metricName == "" {
		http.Error(w, "metrics name is empty", http.StatusNotFound)
		return
	}
	switch metricType {
	case Gauge:
		if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
			storage.gauge[metricName] = convertedMetricValueFloat
		} else {
			w.WriteHeader(http.StatusBadRequest)
			//http.Error(w, "wrong metrics value", http.StatusBadRequest)
			return
		}
	case Counter:
		if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
			if _, ok := storage.counter[metricName]; ok {
				storage.counter[metricName] += int64(convertedMetricValueInt)
			} else {
				storage.counter[metricName] = int64(convertedMetricValueInt)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			//http.Error(w, "wrong metrics value", http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		//http.Error(w, "wrong metrics type", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
