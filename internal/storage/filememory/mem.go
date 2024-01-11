// Package filememory provides an in-memory storage mechanism for metrics data.
package filememory

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
)

// MemStorage represents an in-memory storage structure for metrics data.
type MemStorage struct {
	GaugeMu   sync.Mutex
	Gauge     map[string]float64 `json:"gauge"`
	CounterMu sync.Mutex
	Counter   map[string]int64 `json:"counter"`
}

// NewMemStorage initializes a new instance of MemStorage. It optionally loads existing data
// from a file and sets up a routine for periodic data backup based on server configuration.
//
// Parameters:
// - serverConfigEnable: Boolean indicating whether server configuration is enabled.
// - configuration: A pointer to the server configuration.
//
// Returns:
// - An instance of *MemStorage.
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

// lockCounter locks the mutex for safe concurrent access to the Counter map.
func (s *MemStorage) lockCounter() {
	s.CounterMu.Lock()
}

// lockGauge locks the mutex for safe concurrent access to the Gauge map.
func (s *MemStorage) lockGauge() {
	s.GaugeMu.Lock()
}

// unlockCounter unlocks the mutex for the Counter map.
func (s *MemStorage) unlockCounter() {
	s.CounterMu.Unlock()
}

// unlockGauge unlocks the mutex for the Gauge map.
func (s *MemStorage) unlockGauge() {
	s.GaugeMu.Unlock()
}

// UpdateCounter updates the value of a named counter metric in the storage.
func (s *MemStorage) UpdateCounter(name string, value int64, ok bool) error {
	s.lockCounter()
	defer s.unlockCounter()
	s.Counter[name] += value
	return nil
}

// UpdateGauge updates the value of a named gauge metric in the storage.
func (s *MemStorage) UpdateGauge(name string, value float64) error {
	s.lockGauge()
	defer s.unlockGauge()
	s.Gauge[name] = value
	return nil
}

// GetCounter retrieves the value of a named counter metric from the storage.
func (s *MemStorage) GetCounter(name string) (int64, bool, error) {
	s.lockCounter()
	defer s.unlockCounter()
	value, ok := s.Counter[name]
	return value, ok, nil
}

// GetGauge retrieves the value of a named gauge metric from the storage.
func (s *MemStorage) GetGauge(name string) (float64, bool, error) {
	s.lockGauge()
	defer s.unlockGauge()
	value, ok := s.Gauge[name]
	return value, ok, nil
}

// GetMetrics returns all stored counter and gauge metrics.
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

// generateCombinedData combines gauge and counter metrics into a single map.
func generateCombinedData(s *MemStorage) map[string]interface{} {
	return map[string]interface{}{
		config.Gauge:   s.Gauge,
		config.Counter: s.Counter,
	}
}
