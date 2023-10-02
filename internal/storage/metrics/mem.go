package metrics

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type MetricsData struct {
	MetricsC map[string]int64
	MetricsG map[string]float64
}

type MemStorage struct {
	GaugeMu   sync.Mutex
	Gauge     map[string]float64 `json:"gauge"`
	CounterMu sync.Mutex
	Counter   map[string]int64 `json:"counter"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (st *MemStorage) LockCounter() {
	st.CounterMu.Lock()
}
func (st *MemStorage) LockGauge() {
	st.GaugeMu.Lock()
}

func (st *MemStorage) UnlockCounter() {
	st.CounterMu.Unlock()
}
func (st *MemStorage) UnlockGauge() {
	st.GaugeMu.Unlock()
}

func (st *MemStorage) UpdateCounter(name string, value int64, ok bool) {
	st.LockCounter()
	defer st.UnlockCounter()
	if ok {
		st.Counter[name] += value
	} else {
		st.Counter[name] = value
	}
}

func (st *MemStorage) UpdateGauge(name string, value float64) {
	st.LockGauge()
	defer st.UnlockGauge()
	st.Gauge[name] = value
}

func (st *MemStorage) GetCounter(name string) (int64, bool) {
	st.LockCounter()
	defer st.UnlockCounter()
	value, ok := st.Counter[name]
	return value, ok
}

func (st *MemStorage) GetGauge(name string) (float64, bool) {
	st.LockGauge()
	defer st.UnlockGauge()
	value, ok := st.Gauge[name]
	return value, ok
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
