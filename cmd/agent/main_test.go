package main

import (
	"github.com/elina-chertova/metrics-alerting.git/cmd/storage"
	"runtime"
	"testing"
)

func Test_extractMetrics(t *testing.T) {
	var m runtime.MemStats

	s := &storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}

	t.Run("MemStorage is update", func(t *testing.T) {
		extractMetrics(s, m)
		if len(s.Gauge) == 0 || len(s.Counter) == 0 {
			t.Errorf("Object MemStorage can't be empty")
		}
	})
}
