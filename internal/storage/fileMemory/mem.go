package fileMemory

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"math/rand"
	"runtime"
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

func generateCombinedData(s *MemStorage) map[string]interface{} {
	return map[string]interface{}{
		storage.Gauge:   s.Gauge,
		storage.Counter: s.Counter,
	}
}

func ExtractMetrics(s *MemStorage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	metricsGauge := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"TotalAlloc":    float64(m.TotalAlloc),
		"Sys":           float64(m.Sys),
		"Lookups":       float64(m.Lookups),
		"Mallocs":       float64(m.Mallocs),
		"Frees":         float64(m.Frees),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapSys":       float64(m.HeapSys),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapObjects":   float64(m.HeapObjects),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"BuckHashSys":   float64(m.BuckHashSys),
		"GCSys":         float64(m.GCSys),
		"OtherSys":      float64(m.OtherSys),
		"NextGC":        float64(m.NextGC),
		"LastGC":        float64(m.LastGC),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"NumGC":         float64(m.NumGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"GCCPUFraction": float64(m.GCCPUFraction),
		"RandomValue":   generator.Float64(),
	}
	for name, value := range metricsGauge {
		s.UpdateGauge(name, value)
	}
	s.UpdateCounter("PollCount", 1, true)
}
