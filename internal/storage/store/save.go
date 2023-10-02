package store

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
)

func (s *storager) Save(fileName string) {
	combinedData := map[string]interface{}{
		"gauge":   s.memStorage.Gauge,
		"counter": s.memStorage.Counter,
	}

	data, err := json.MarshalIndent(combinedData, "", "   ")
	if err != nil {
		fmt.Printf("error MarshalIndent: %v\n", err)
	}
	fmt.Println("here", string(data))
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0777); err != nil {
		fmt.Printf("error creating file: %v\n", err)
	}
	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		fmt.Printf("error creating JSON: %v\n", err)
	}

}
