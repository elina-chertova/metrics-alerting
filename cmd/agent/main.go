package main

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/cmd/agent/flags"
	"github.com/elina-chertova/metrics-alerting.git/cmd/storage"
	"github.com/levigross/grequests"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var url_update = "http://" + flags.FlagAddress + "/update"

func extractMetrics(s *storage.MemStorage, m runtime.MemStats) {
	runtime.ReadMemStats(&m)
	s.GaugeMu.Lock()
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))

	s.Gauge["Alloc"] = float64(m.Alloc)
	s.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	s.Gauge["Sys"] = float64(m.Sys)
	s.Gauge["Lookups"] = float64(m.Lookups)
	s.Gauge["Mallocs"] = float64(m.Mallocs)
	s.Gauge["Frees"] = float64(m.Frees)
	s.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	s.Gauge["HeapSys"] = float64(m.HeapSys)
	s.Gauge["HeapIdle"] = float64(m.HeapIdle)
	s.Gauge["HeapInuse"] = float64(m.HeapInuse)
	s.Gauge["HeapReleased"] = float64(m.HeapReleased)
	s.Gauge["HeapObjects"] = float64(m.HeapObjects)
	s.Gauge["StackInuse"] = float64(m.StackInuse)
	s.Gauge["StackSys"] = float64(m.StackSys)
	s.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
	s.Gauge["MSpanSys"] = float64(m.MSpanSys)
	s.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	s.Gauge["MCacheSys"] = float64(m.MCacheSys)
	s.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	s.Gauge["GCSys"] = float64(m.GCSys)
	s.Gauge["OtherSys"] = float64(m.OtherSys)
	s.Gauge["NextGC"] = float64(m.NextGC)
	s.Gauge["LastGC"] = float64(m.LastGC)
	s.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.Gauge["NumGC"] = float64(m.NumGC)
	s.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	s.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	s.Gauge["RandomValue"] = generator.Float64()
	s.GaugeMu.Unlock()

	s.CounterMu.Lock()
	s.Counter["PollCount"] += 1
	s.CounterMu.Unlock()
}

func sendRequest(name string, value any, metricsType string) {
	metricURL := fmt.Sprintf("%s/%s/%s/%v", url_update, metricsType, name, value)
	_, err := grequests.Post(metricURL, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}
}

func Requests(s *storage.MemStorage) {

	var wg sync.WaitGroup

	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue any) {
			defer wg.Done()
			sendRequest(metricName, metricValue, "gauge")
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue any) {
			defer wg.Done()
			sendRequest(metricName, metricValue, "counter")
		}(metricName, metricValue)
	}
	wg.Wait()
}

func main() {
	flags.ParseAgentFlags()
	fmt.Println("Running server on", flags.PollInterval)
	var mem runtime.MemStats
	storage := &storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	go func() {
		for {
			extractMetrics(storage, mem)
			time.Sleep(flags.PollInterval)
		}
	}()
	go func() {
		for {
			time.Sleep(flags.ReportInterval)
			Requests(storage)
		}
	}()

	select {}
}
