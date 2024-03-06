package request

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
)

func TestCompressData(t *testing.T) {
	data := []byte("test data for compression")
	compressedData := compressData(data)

	var uncompressedData bytes.Buffer
	gzipReader, err := gzip.NewReader(&compressedData)
	if err != nil {
		t.Fatalf("Error creating gzip reader: %v", err)
	}
	if _, err := uncompressedData.ReadFrom(gzipReader); err != nil {
		t.Fatalf("Error decompressing data: %v", err)
	}
	if uncompressedData.String() != string(data) {
		t.Errorf(
			"Compressed data did not match original, got: %s, want: %s",
			uncompressedData.String(),
			data,
		)
	}
}

func TestFormJSON(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      interface{}
		metricType string
		wantErr    bool
	}{
		{"Valid Gauge", "testGauge", 1.23, config.Gauge, false},
		{"Valid Counter", "testCounter", int64(123), config.Counter, false},
		{"Invalid Type", "testMetric", 123, "invalid", true},
		{"Invalid Value Type", "testMetric", "invalid", config.Gauge, true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := formJSON(tt.metricName, tt.value, tt.metricType)
				if (err != nil) != tt.wantErr {
					t.Errorf("formJSON() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
