package metrics

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
)

func (s *MemStorage) Load(fileName string) {
	combinedData := GenerateCombinedData(s)
	dataNew, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("error reading JSON: %v\n", err)
	}

	if err := json.Unmarshal(dataNew, &combinedData); err != nil {
		fmt.Printf("error Unmarshal JSON: %v\n", err)
	}
	s.UpdateBackupMap(combinedData)
}
