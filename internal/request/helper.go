package request

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/levigross/grequests"
)

var (
	ErrHTTPRequestCreation   = errors.New("error creating HTTP request")
	ErrHTTPRequestFailed     = errors.New("HTTP request failed")
	ErrCompressData          = errors.New("error compressing data")
	ErrUnsupportedValueType  = errors.New("unsupported value type")
	ErrUnsupportedMetricType = errors.New("unsupported metric type")
)

type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e RetryableError) Retryable() bool {
	return true
}

func sendRequest(contentType string, isCompress bool, url string, jsonBody []byte) error {
	var ro *grequests.RequestOptions
	if isCompress {
		compressedData := compressData(jsonBody)

		ro = &grequests.RequestOptions{
			Headers: map[string]string{
				"Content-Type":     contentType,
				"Content-Encoding": "gzip",
			},
			RequestBody: &compressedData,
		}
	} else {
		ro = &grequests.RequestOptions{
			Headers: map[string]string{"Content-Type": contentType},
			JSON:    jsonBody,
		}
	}

	resp, err := grequests.Post(url, ro)
	if err != nil {
		return RetryableError{Err: ErrHTTPRequestCreation}
	}
	defer resp.Close()
	if !resp.Ok {
		return RetryableError{
			Err: fmt.Errorf(
				"%w with status code %d",
				ErrHTTPRequestFailed,
				resp.StatusCode,
			),
		}
	}
	return nil
}

func compressData(data []byte) bytes.Buffer {
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)

	_, err := gzipWriter.Write(data)
	if err != nil {
		fmt.Printf("error compressing data %s", ErrCompressData.Error())
	}
	gzipWriter.Close()
	return compressedBuffer
}

func formJSON(metricName string, value interface{}, typeMetric string) (f.Metric, error) {
	var metrics f.Metric

	switch typeMetric {
	case config.Gauge:
		var v float64
		switch value := value.(type) {
		case int64:
			v = float64(value)
		case float64:
			v = value
		default:
			return f.Metric{}, ErrUnsupportedValueType
		}
		metrics = f.Metric{
			ID:    metricName,
			MType: config.Gauge,
			Value: &v,
		}
	case config.Counter:
		var delta int64
		switch value := value.(type) {
		case int64:
			delta = value
		case float64:
			delta = int64(value)
		default:
			return f.Metric{}, ErrUnsupportedValueType
		}
		metrics = f.Metric{
			ID:    metricName,
			MType: config.Counter,
			Delta: &delta,
		}
	default:
		return f.Metric{}, ErrUnsupportedMetricType
	}

	return metrics, nil
}
