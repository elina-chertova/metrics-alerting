// Package filememory provides functionality for in-memory storage of metrics data
// and supports operations like data backup to a file.
package filememory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/goccy/go-json"
)

// BackupError represents an error that occurs during the backup process.
type BackupError struct {
	Message string
	Err     error
}

// Error returns a formatted string combining the error message and the underlying error.
func (e BackupError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// backup performs a backup of the metrics data stored in memory to a specified file.
//
// Parameters:
// - fileName: The name of the file where the backup data will be stored.
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
