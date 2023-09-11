package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	MethodPost = "POST"

	Gauge   = "gauge"
	Counter = "counter"
)

type MemStorage struct {
	gaugeMu   sync.Mutex
	gauge     map[string]float64
	counterMu sync.Mutex
	counter   map[string]int64
}

type Middleware func(http.Handler) http.Handler

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

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
func (storage *MemStorage) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	switch metricType {
	case Gauge:
		if convertedMetricValueFloat, err := strconv.ParseFloat(metricValue, 64); err == nil {
			storage.gaugeMu.Lock()
			storage.gauge[metricName] = convertedMetricValueFloat
			storage.gaugeMu.Unlock()
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case Counter:
		if convertedMetricValueInt, err := strconv.Atoi(metricValue); err == nil {
			storage.counterMu.Lock()
			_, ok := storage.counter[metricName]
			if ok {
				storage.counter[metricName] += int64(convertedMetricValueInt)
			} else {
				storage.counter[metricName] = int64(convertedMetricValueInt)
			}
			storage.counterMu.Unlock()
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
