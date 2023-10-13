package filememory

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/goccy/go-json"
	"os"
)

func (s *MemStorage) load(fileName string) {
	combinedData := generateCombinedData(s)
	dataNew, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("error reading JSON: %v\n", err)
	}

	if err := json.Unmarshal(dataNew, &combinedData); err != nil {
		fmt.Printf("error Unmarshal JSON: %v\n", err)
	}
	s.updateBackupMap(combinedData)
}

func (s *MemStorage) updateBackupMap(combinedData map[string]interface{}) {
	for key, value := range combinedData {
		switch key {
		case storage.Gauge:
			if gaugeData, ok := value.(map[string]interface{}); ok {
				gauge := make(map[string]float64)
				for k, v := range gaugeData {
					if floatValue, isFloat := v.(float64); isFloat {
						gauge[k] = floatValue
					}
				}
				s.Gauge = gauge
			}

		case storage.Counter:
			if counterData, ok := value.(map[string]interface{}); ok {
				counter := make(map[string]int64)
				for k, v := range counterData {
					if floatValue, isFloat := v.(float64); isFloat {
						counter[k] = int64(floatValue)
					}
				}
				s.Counter = counter
			}
		}
	}
}
