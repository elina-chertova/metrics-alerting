package store

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/metrics"
	"github.com/goccy/go-json"
	"os"
)

func (s *storager) Load(fileName string) {
	combinedData := map[string]interface{}{
		"gauge":   s.memStorage.Gauge,
		"counter": s.memStorage.Counter,
	}
	dataNew, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("error reading JSON: %v\n", err)
	}

	if err := json.Unmarshal(dataNew, &combinedData); err != nil {
		fmt.Printf("error Unmarshal JSON: %v\n", err)
	}
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
