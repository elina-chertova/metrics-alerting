package filememory

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
)

func TestLoadFunction(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "example.*.txt")
	if err != nil {
		t.Fatalf("Cannot create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	testData := `{"gauge":{"testGauge":1.23},"counter":{"testCounter":123}}`
	if _, err := tmpfile.Write([]byte(testData)); err != nil {
		t.Fatalf("Cannot write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Cannot close temporary file: %v", err)
	}

	storage := NewMemStorage(false, nil)
	storage.load(tmpfile.Name())
	if val, ok := storage.Gauge["testGauge"]; !ok || val != 1.23 {
		t.Errorf("Gauge not updated correctly: got %v, want 1.23", val)
	}
	if val, ok := storage.Counter["testCounter"]; !ok || val != 123 {
		t.Errorf("Counter not updated correctly: got %v, want 123", val)
	}
}

func TestUpdateBackupMap(t *testing.T) {
	storage := &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}

	combinedData := map[string]interface{}{
		config.Gauge: map[string]interface{}{
			"testGauge": 2.34,
		},
		config.Counter: map[string]interface{}{
			"testCounter": 456.0,
		},
	}

	storage.updateBackupMap(combinedData)

	expectedGaugeValue := 2.34
	if val, ok := storage.Gauge["testGauge"]; !ok || val != expectedGaugeValue {
		t.Errorf("Gauge not updated correctly: got %v, want %v", val, expectedGaugeValue)
	}

	expectedCounterValue := int64(456)
	if val, ok := storage.Counter["testCounter"]; !ok || val != expectedCounterValue {
		t.Errorf("Counter not updated correctly: got %v, want %v", val, expectedCounterValue)
	}
}

func TestLoadError_Error(t *testing.T) {
	testErr := errors.New("test error")
	loadErr := LoadError{
		Message: "An error occurred",
		Err:     testErr,
	}

	expected := fmt.Sprintf("An error occurred: %v", testErr)
	actual := loadErr.Error()

	if actual != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, actual)
	}
}
