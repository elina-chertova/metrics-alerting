package filememory

import (
	"errors"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"sync"
	"time"
)

type MemStorage struct {
	GaugeMu   sync.Mutex
	Gauge     map[string]float64 `json:"gauge"`
	CounterMu sync.Mutex
	Counter   map[string]int64 `json:"counter"`
}

func NewMemStorage(serverConfigEnable bool, configuration *config.Server) *MemStorage {
	s := &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	if serverConfigEnable {
		if configuration.FlagRestore {
			s.load(configuration.FileStoragePath)
		}
		go func() {
			for {
				time.Sleep(time.Duration(configuration.StoreInterval) * time.Second)
				s.backup(configuration.FileStoragePath)
			}
		}()
	}
	return s
}

func (s *MemStorage) LockCounter() {
	s.CounterMu.Lock()
}
func (s *MemStorage) LockGauge() {
	s.GaugeMu.Lock()
}

func (s *MemStorage) UnlockCounter() {
	s.CounterMu.Unlock()
}
func (s *MemStorage) UnlockGauge() {
	s.GaugeMu.Unlock()
}

func (s *MemStorage) UpdateCounter(name string, value int64, ok bool) {
	s.LockCounter()
	defer s.UnlockCounter()
	if ok {
		s.Counter[name] += value
	} else {
		s.Counter[name] = value
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.LockGauge()
	defer s.UnlockGauge()
	s.Gauge[name] = value
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.LockCounter()
	defer s.UnlockCounter()
	value, ok := s.Counter[name]
	return value, ok
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.LockGauge()
	defer s.UnlockGauge()
	value, ok := s.Gauge[name]
	return value, ok
}

func (s *MemStorage) GetMetrics() (map[string]int64, map[string]float64) {
	s.LockGauge()
	s.LockCounter()
	defer s.UnlockGauge()
	defer s.UnlockCounter()
	return s.Counter, s.Gauge
}

var ErrNotAllowed = errors.New("method not allowed")

func (s *MemStorage) InsertBatchMetrics(metrics []f.Metric) error {
	return fmt.Errorf("%w", ErrNotAllowed)
}

func generateCombinedData(s *MemStorage) map[string]interface{} {
	return map[string]interface{}{
		config.Gauge:   s.Gauge,
		config.Counter: s.Counter,
	}
}
