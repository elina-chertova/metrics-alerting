package filememory

import (
	"errors"
	"reflect"
	"testing"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
)

func TestNewMemStorage(t *testing.T) {
	conf := &config.Server{}

	storage := NewMemStorage(false, conf)
	if storage == nil {
		t.Errorf("NewMemStorage returned nil")
	} else if len(storage.Gauge) != 0 || len(storage.Counter) != 0 {
		t.Errorf("NewMemStorage should initialize empty maps")
	}
}

func TestInsertBatchMetrics(t *testing.T) {
	storage := &MemStorage{}

	err := storage.InsertBatchMetrics(nil)
	if err == nil {
		t.Errorf("InsertBatchMetrics should return an error")
	}
	if !errors.Is(err, ErrNotAllowed) {
		t.Errorf("InsertBatchMetrics should return ErrNotAllowed error")
	}
}

func TestUpdateCounter(t *testing.T) {
	s := NewMemStorage(false, nil)
	s.UpdateCounter("TestCounter", 42, true)
	value, ok, _ := s.GetCounter("TestCounter")

	if value != 42 || !ok {
		t.Errorf("UpdateCounter or GetCounter didn't work as expected")
	}
}

func TestUpdateGauge(t *testing.T) {
	s := NewMemStorage(false, nil)
	s.UpdateGauge("TestGauge", 3.14)
	value, ok, _ := s.GetGauge("TestGauge")

	if value != 3.14 || !ok {
		t.Errorf("UpdateGauge or GetGauge didn't work as expected")
	}
}

func TestGetMetrics(t *testing.T) {
	s := NewMemStorage(false, nil)
	s.Gauge["hello"] = 4.3
	s.Counter["count"] = 6
	c, g := s.GetMetrics()

	value1, exists1 := c["count"]
	value2, exists2 := g["hello"]

	if value1 != 6 || !exists1 || value2 != 4.3 || !exists2 {
		t.Errorf("GetMetrics didn't work as expected")
	}
}

func TestGenerateCombinedData(t *testing.T) {
	storage := &MemStorage{
		Gauge: map[string]float64{
			"gauge1": 1.23,
			"gauge2": 4.56,
		},
		Counter: map[string]int64{
			"counter1": 123,
			"counter2": 456,
		},
	}

	combinedData := generateCombinedData(storage)

	expectedResult := map[string]interface{}{
		config.Gauge:   storage.Gauge,
		config.Counter: storage.Counter,
	}

	if !reflect.DeepEqual(combinedData, expectedResult) {
		t.Errorf("generateCombinedData() = %v, want %v", combinedData, expectedResult)
	}
}
