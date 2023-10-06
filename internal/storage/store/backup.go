package store

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
)

func (s *storager) Backup(fileName string) {
	combinedData := GenerateCombinedData(s)
	data, err := json.MarshalIndent(combinedData, "", "   ")
	if err != nil {
		fmt.Printf("error MarshalIndent: %v\n", err)
	}

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		dir := filepath.Dir(fileName)
		if err := os.MkdirAll(dir, 0777); err != nil {
			fmt.Printf("error creating file: %v\n", err)
		}
	}

	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		fmt.Printf("error creating JSON: %v\n", err)
	}

}