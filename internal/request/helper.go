// Package request handles the sending of metrics data to a specified endpoint.
package request

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/security"
	"github.com/levigross/grequests"
)

var (
	ErrHTTPRequestCreation   = errors.New("error creating HTTP request")
	ErrHTTPRequestFailed     = errors.New("HTTP request failed")
	ErrCompressData          = errors.New("error compressing data")
	ErrUnsupportedValueType  = errors.New("unsupported value type")
	ErrUnsupportedMetricType = errors.New("unsupported metric type")
)

// RetryableError represents an error that occurred during an HTTP request that may be retried.
type RetryableError struct {
	Err error
}

// Error returns a formatted error message indicating the error is retryable.
func (e RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

// Retryable indicates whether the error is retryable.
func (e RetryableError) Retryable() bool {
	return true
}

// sendRequest creates and sends an HTTP request with optional compression and security hashing.
//
// Parameters:
// - contentType: The MIME type of the content.
// - isCompress: Indicates whether the data should be compressed.
// - url: The URL to which the request is sent.
// - jsonBody: The JSON formatted data to be sent.
// - secretKey: A secret key used for generating a hash (optional).
//
// Returns:
// - A RetryableError if the request creation or execution fails.
func sendRequest(
	contentType string,
	isCompress bool,
	url string,
	jsonBody []byte,
	secretKey string,
) error {
	var ro *grequests.RequestOptions
	headers := make(map[string]string)
	headers["Content-Type"] = contentType

	if secretKey != "" {
		hashBody := security.Hash(string(jsonBody), []byte(secretKey))
		headers["HashSHA256"] = hashBody
	}

	if isCompress {
		compressedData := compressData(jsonBody)
		headers["Content-Encoding"] = "gzip"
		ro = &grequests.RequestOptions{
			Headers:     headers,
			RequestBody: &compressedData,
		}
	} else {
		ro = &grequests.RequestOptions{
			Headers: headers,
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

// compressData compresses the given data using gzip compression.
//
// Parameters:
// - data: The data to be compressed.
//
// Returns:
// - A bytes.Buffer containing the compressed data.
func compressData(data []byte) bytes.Buffer {
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)

	_, err := gzipWriter.Write(data)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("error compressing data %s", ErrCompressData.Error()))
	}
	gzipWriter.Close()
	return compressedBuffer
}

// formJSON creates a formatter.Metric from a metric name, value, and type.
//
// Parameters:
// - metricName: The name of the metric.
// - value: The value of the metric.
// - typeMetric: The type of the metric (gauge or counter).
//
// Returns:
// - A formatter.Metric instance and an error if the metric type or value type is unsupported.
func formJSON(metricName string, value interface{}, typeMetric string) (formatter.Metric, error) {
	var metrics formatter.Metric

	switch typeMetric {
	case config.Gauge:
		var v float64
		switch value := value.(type) {
		case int64:
			v = float64(value)
		case float64:
			v = value
		default:
			return formatter.Metric{}, ErrUnsupportedValueType
		}
		metrics = formatter.Metric{
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
			return formatter.Metric{}, ErrUnsupportedValueType
		}
		metrics = formatter.Metric{
			ID:    metricName,
			MType: config.Counter,
			Delta: &delta,
		}
	default:
		return formatter.Metric{}, ErrUnsupportedMetricType
	}

	return metrics, nil
}
