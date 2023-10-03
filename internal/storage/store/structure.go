package store

import "github.com/elina-chertova/metrics-alerting.git/internal/storage/metrics"

type storager struct{ memStorage *metrics.MemStorage }

func NewStorager(s *metrics.MemStorage) *storager {
	return &storager{s}
}

func GenerateCombinedData(s *storager) map[string]interface{} {
	return map[string]interface{}{
		metrics.Gauge:   s.memStorage.Gauge,
		metrics.Counter: s.memStorage.Counter,
	}
}

func (s *storager) UpdateBackupMap(combinedData map[string]interface{}) {
	for key, value := range combinedData {
		switch key {
		case metrics.Gauge:
			if gaugeData, ok := value.(map[string]interface{}); ok {
				gauge := make(map[string]float64)
				for k, v := range gaugeData {
					if floatValue, isFloat := v.(float64); isFloat {
						gauge[k] = floatValue
					}
				}
				s.memStorage.Gauge = gauge
			}

		case metrics.Counter:
			if counterData, ok := value.(map[string]interface{}); ok {
				counter := make(map[string]int64)
				for k, v := range counterData {
					if floatValue, isFloat := v.(float64); isFloat {
						counter[k] = int64(floatValue)
					}
				}
				s.memStorage.Counter = counter
			}
		}
	}
}
