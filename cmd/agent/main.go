package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type MemStorage struct {
	gaugeMu   sync.RWMutex
	gauge     map[string]float64
	counterMu sync.RWMutex
	counter   map[string]int64
}

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	url            = "http://localhost:8080/update"
)

func extractMetrics(s *MemStorage, m runtime.MemStats) {
	runtime.ReadMemStats(&m)
	s.gaugeMu.Lock()
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	s.gauge["Alloc"] = float64(m.Alloc)
	s.gauge["TotalAlloc"] = float64(m.TotalAlloc)
	s.gauge["Sys"] = float64(m.Sys)
	s.gauge["Lookups"] = float64(m.Lookups)
	s.gauge["Mallocs"] = float64(m.Mallocs)
	s.gauge["Frees"] = float64(m.Frees)
	s.gauge["HeapAlloc"] = float64(m.HeapAlloc)
	s.gauge["HeapSys"] = float64(m.HeapSys)
	s.gauge["HeapIdle"] = float64(m.HeapIdle)
	s.gauge["HeapInuse"] = float64(m.HeapInuse)
	s.gauge["HeapReleased"] = float64(m.HeapReleased)
	s.gauge["HeapObjects"] = float64(m.HeapObjects)
	s.gauge["StackInuse"] = float64(m.StackInuse)
	s.gauge["StackSys"] = float64(m.StackSys)
	s.gauge["MSpanInuse"] = float64(m.MSpanInuse)
	s.gauge["MSpanSys"] = float64(m.MSpanSys)
	s.gauge["MCacheInuse"] = float64(m.MCacheInuse)
	s.gauge["MCacheSys"] = float64(m.MCacheSys)
	s.gauge["BuckHashSys"] = float64(m.BuckHashSys)
	s.gauge["GCSys"] = float64(m.GCSys)
	s.gauge["OtherSys"] = float64(m.OtherSys)
	s.gauge["NextGC"] = float64(m.NextGC)
	s.gauge["LastGC"] = float64(m.LastGC)
	s.gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.gauge["NumGC"] = float64(m.NumGC)
	s.gauge["NumForcedGC"] = float64(m.NumForcedGC)
	s.gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	s.gauge["RandomValue"] = generator.Float64()
	s.gaugeMu.Unlock()

	s.counterMu.Lock()
	s.counter["PollCount"] += 1
	s.counterMu.Unlock()
}

func sendRequest(name string, value any, metricsType string) {
	metricURL := fmt.Sprintf("%s/%s/%s/%v", url, metricsType, name, value)
	body, err := http.Post(metricURL, "text/plain", nil)
	defer body.Body.Close()
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}
}

func Requests(s *MemStorage) {

	var wg sync.WaitGroup

	for metricName, metricValue := range s.gauge {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendRequest(metricName, metricValue, "gauge")
		}()
	}

	for metricName, metricValue := range s.counter {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendRequest(metricName, metricValue, "counter")
		}()
	}
	wg.Wait()
}

func main() {
	var mem runtime.MemStats

	storage := &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
	go func() {
		for {
			extractMetrics(storage, mem)
			time.Sleep(pollInterval)
		}
	}()
	go func() {
		for {
			time.Sleep(reportInterval)
			Requests(storage)
		}
	}()

	select {}
}
