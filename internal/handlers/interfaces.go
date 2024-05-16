package handlers

import (
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
)

// MetricsStorage defines an interface for storing and retrieving metric data.
type MetricsStorage interface {
	UpdateCounter(name string, value int64, ok bool) error
	UpdateGauge(name string, value float64) error
	GetCounter(name string) (int64, bool, error)
	GetGauge(name string) (float64, bool, error)
	GetMetrics() (map[string]int64, map[string]float64)
	InsertBatchMetrics([]f.Metric) error
}
