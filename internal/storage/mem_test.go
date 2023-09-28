package storage

import (
	"testing"
)

func TestExtractMetrics(t *testing.T) {
	st := &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	ExtractMetrics(st)

	if len(st.Counter) <= 0 || len(st.Gauge) <= 0 {
		t.Errorf("Expected the length of st to be greater than 0, but got len(st.Counter)=%d, len(st.Gauge)=%d", len(st.Counter), len(st.Gauge))
	}
}
