package request

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/levigross/grequests"
)

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
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	defer resp.Close()

	if !resp.Ok {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	return nil
}

func compressData(data []byte) bytes.Buffer {
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)

	_, err := gzipWriter.Write(data)
	if err != nil {
		_ = fmt.Errorf("error compressing data: %v", err)
	}
	gzipWriter.Close()
	return compressedBuffer
}

func formJSON(metricName string, value any, typeMetric string) f.Metric {
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
			_ = fmt.Errorf("unsupported value type: %T", value)
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
			_ = fmt.Errorf("unsupported value type: %T", value)
		}
		metrics = f.Metric{
			ID:    metricName,
			MType: config.Counter,
			Delta: &delta,
		}
	default:
		_ = fmt.Errorf("unsupported metric type: %s", typeMetric)
	}

	return metrics
}
