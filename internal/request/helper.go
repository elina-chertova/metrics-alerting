package request

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
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

type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e RetryableError) Retryable() bool {
	return true
}

func sendRequest(
	contentType string,
	isCompress bool,
	url string,
	jsonBody []byte,
	secretKey string,
) error {
	var ro *grequests.RequestOptions
	var headers map[string]string

	headers = make(map[string]string)
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
