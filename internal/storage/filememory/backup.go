package filememory

import (
	"errors"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/goccy/go-json"
	"os"
	"path/filepath"
)

type BackupError struct {
	Message string
	Err     error
}

func (e BackupError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (s *MemStorage) backup(fileName string) {
	combinedData := generateCombinedData(s)
	data, err := json.MarshalIndent(combinedData, "", "   ")
	if err != nil {
		logger.Log.Error(BackupError{Err: err, Message: "failed to marshal data"}.Error())
	}

	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		dir := filepath.Dir(fileName)
		if err := os.MkdirAll(dir, 0777); err != nil {
			logger.Log.Error(BackupError{Err: err, Message: "failed to create directory"}.Error())
		}
	}

	err = os.WriteFile(fileName, data, 0666)
	if err != nil {
		logger.Log.Error(BackupError{Err: err, Message: "failed to write data"}.Error())
	}
}
