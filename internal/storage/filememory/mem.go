package filememory

import (
	"errors"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
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

func (s *MemStorage) lockCounter() {
	s.CounterMu.Lock()
}
func (s *MemStorage) lockGauge() {
	s.GaugeMu.Lock()
}

func (s *MemStorage) unlockCounter() {
	s.CounterMu.Unlock()
}
func (s *MemStorage) unlockGauge() {
	s.GaugeMu.Unlock()
}

func (s *MemStorage) UpdateCounter(name string, value int64, ok bool) error {
	s.lockCounter()
	defer s.unlockCounter()
	if ok {
		s.Counter[name] += value
	} else {
		s.Counter[name] = value
	}
	return nil
}

func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.lockGauge()
	defer s.unlockGauge()
	s.Gauge[name] = value
	return nil
}

func (s *MemStorage) GetCounter(name string) (int64, bool, error) {
	s.lockCounter()
	defer s.unlockCounter()
	value, ok := s.Counter[name]
	return value, ok, nil
}

func (s *MemStorage) GetGauge(name string) (float64, bool, error) {
	s.lockGauge()
	defer s.unlockGauge()
	value, ok := s.Gauge[name]
	return value, ok, nil
}

func (s *MemStorage) GetMetrics() (map[string]int64, map[string]float64) {
	s.lockGauge()
	s.lockCounter()
	defer s.unlockGauge()
	defer s.unlockCounter()
	return s.Counter, s.Gauge
}

var ErrNotAllowed = errors.New("method not allowed")

func (s *MemStorage) InsertBatchMetrics(metrics []formatter.Metric) error {
	return fmt.Errorf("%w", ErrNotAllowed)
}

func generateCombinedData(s *MemStorage) map[string]interface{} {
	return map[string]interface{}{
		config.Gauge:   s.Gauge,
		config.Counter: s.Counter,
	}
}
