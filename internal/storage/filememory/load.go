package filememory

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/goccy/go-json"
	"os"
)

type LoadError struct {
	Message string
	Err     error
}

func (e LoadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (s *MemStorage) load(fileName string) {
	combinedData := generateCombinedData(s)
	dataNew, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Print(LoadError{Err: err, Message: "failed to read data from file"}.Error())
	}

	if err := json.Unmarshal(dataNew, &combinedData); err != nil {
		fmt.Print(LoadError{Err: err, Message: "failed to unmarshal JSON"}.Error())
	}
	s.updateBackupMap(combinedData)
}

func (s *MemStorage) updateBackupMap(combinedData map[string]interface{}) {
	for key, value := range combinedData {
		switch key {
		case config.Gauge:
			if gaugeData, ok := value.(map[string]interface{}); ok {
				gauge := make(map[string]float64)
				for k, v := range gaugeData {
					if floatValue, isFloat := v.(float64); isFloat {
						gauge[k] = floatValue
					}
				}
				s.Gauge = gauge
			}

		case config.Counter:
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
