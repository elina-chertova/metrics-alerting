package main

import (
	"runtime"
	"testing"
)

func Test_extractMetrics(t *testing.T) {
	var m runtime.MemStats

	s := &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}

	t.Run("MemStorage is update", func(t *testing.T) {
		extractMetrics(s, m)
		if len(s.gauge) == 0 || len(s.counter) == 0 {
			t.Errorf("Object MemStorage can't be empty")
		}
	})
}
