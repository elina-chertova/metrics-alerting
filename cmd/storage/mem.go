package storage

import "sync"

type MetricsData struct {
	MetricsC map[string]int64
	MetricsG map[string]float64
}

type MemStorage struct {
	GaugeMu   sync.Mutex
	Gauge     map[string]float64 `json:"gauge"`
	CounterMu sync.Mutex
	Counter   map[string]int64 `json:"counter"`
}
