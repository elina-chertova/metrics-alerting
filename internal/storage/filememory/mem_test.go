package filememory

import (
	"testing"
)

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
