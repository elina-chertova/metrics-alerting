// Package metricextractor contains functions for extracting system and application metrics.
package metricextractor

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// ExtractOSMetrics extracts operating system level metrics such as total memory,
// free memory, and CPU utilization.
//
// Parameters:
// - s: A memory storage instance where the extracted metrics will be stored.
//
// Returns:
// - An error if there is an issue in extracting metrics or updating the storage.
func ExtractOSMetrics(s *filememory.MemStorage) error {
	v, _ := mem.VirtualMemory()
	CPUUtilized, err := cpu.Percent(time.Second, true)
	if err != nil {
		return err
	}
	metricsGauge := map[string]float64{
		"TotalMemory":     float64(v.Total),
		"FreeMemory":      float64(v.Free),
		"CPUutilization1": CPUUtilized[0],
	}

	for name, value := range metricsGauge {
		err := s.UpdateGauge(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExtractMetrics gathers various runtime metrics of the Go application, including
// memory allocation, GC statistics, and other system metrics. It also generates a
// random value as part of the metrics.
//
// Parameters:
// - s: A memory storage instance where the extracted metrics will be stored.
//
// Returns:
// - An error if there is an issue in extracting metrics or updating the storage.
func ExtractMetrics(s *filememory.MemStorage) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	metricsGauge := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"TotalAlloc":    float64(m.TotalAlloc),
		"Sys":           float64(m.Sys),
		"Lookups":       float64(m.Lookups),
		"Mallocs":       float64(m.Mallocs),
		"Frees":         float64(m.Frees),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapSys":       float64(m.HeapSys),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapObjects":   float64(m.HeapObjects),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"BuckHashSys":   float64(m.BuckHashSys),
		"GCSys":         float64(m.GCSys),
		"OtherSys":      float64(m.OtherSys),
		"NextGC":        float64(m.NextGC),
		"LastGC":        float64(m.LastGC),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"NumGC":         float64(m.NumGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"GCCPUFraction": m.GCCPUFraction,
		"RandomValue":   generator.Float64(),
	}
	for name, value := range metricsGauge {
		err := s.UpdateGauge(name, value)
		if err != nil {
			return err
		}
	}
	err := s.UpdateCounter("PollCount", 1, true)
	if err != nil {
		return err
	}
	return nil
}
