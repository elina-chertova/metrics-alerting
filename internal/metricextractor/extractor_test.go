package metricextractor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
)

func TestExtractMetrics(t *testing.T) {
	st := &filememory.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	err := ExtractMetrics(st)
	if err != nil {
		return
	}

	if len(st.Counter) <= 0 || len(st.Gauge) <= 0 {
		t.Errorf(
			"Expected the length of st to be greater than 0, but got len(st.Counter)=%d, len(st.Gauge)=%d",
			len(st.Counter),
			len(st.Gauge),
		)
	}
}

func TestExtractOSMetrics(t *testing.T) {
	st := &filememory.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	err := ExtractOSMetrics(st)
	assert.NoError(t, err)

}
