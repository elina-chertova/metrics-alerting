// Package filememory handles the loading and saving of metrics data to and from files.
package filememory

import (
	"fmt"
	"os"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/goccy/go-json"
)

// LoadError represents an error that occurs during the loading process of metrics data.
type LoadError struct {
	Message string
	Err     error
}

// Error returns a formatted string that combines the error message and the underlying error.
func (e LoadError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// load retrieves metrics data from a specified file and updates the in-memory storage.
//
// Parameters:
// - fileName: The name of the file from which to load the data.
func (s *MemStorage) load(fileName string) {
	combinedData := generateCombinedData(s)
	dataNew, err := os.ReadFile(fileName)
	if err != nil {
		logger.Log.Error(BackupError{Err: err, Message: "failed to read data from file"}.Error())
	}

	if err := json.Unmarshal(dataNew, &combinedData); err != nil {
		logger.Log.Error(BackupError{Err: err, Message: "failed to unmarshal JSON"}.Error())
	}
	s.updateBackupMap(combinedData)
}

// updateBackupMap updates the in-memory storage with data from a combined data map.
//
// Parameters:
// - combinedData: A map containing combined gauge and counter metrics data.
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
