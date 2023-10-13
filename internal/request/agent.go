package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage"
	metrics2 "github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"github.com/levigross/grequests"
	"sync"
)

func MetricsToServer(
	s *metrics2.MemStorage,
	contentType string,
	url string,
	isCompress bool,
) error {
	switch contentType {
	case f.ContentTypeTextPlain:
		return metricsToServerTextPlain(s, url, isCompress)
	case f.ContentTypeJSON:
		return metricsToServerAppJSON(s, url, isCompress)
	default:
		return fmt.Errorf("error creating HTTP request, wrong Content-Type: %s", contentType)
	}
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
	case storage.Gauge:
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
			MType: storage.Gauge,
			Value: &v,
		}
	case storage.Counter:
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
			MType: storage.Counter,
			Delta: &delta,
		}
	default:
		_ = fmt.Errorf("unsupported metric type: %s", typeMetric)
	}

	return metrics
}

func metricsToServerAppJSON(s *metrics2.MemStorage, url string, isCompress bool) error {
	var wg sync.WaitGroup

	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metrics := formJSON(metricName, metricValue, storage.Gauge)
			out, err := json.Marshal(metrics)
			if err != nil {
				fmt.Printf("error creating JSON: %v\n", err)
			}
			if err := sendRequest(f.ContentTypeJSON, isCompress, url, out); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metrics := formJSON(metricName, metricValue, storage.Counter)
			out, err := json.Marshal(metrics)

			if err != nil {
				fmt.Printf("error creating JSON: %v\n", err)
			}
			if err := sendRequest(f.ContentTypeJSON, isCompress, url, out); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}

		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
}

func metricsToServerTextPlain(s *metrics2.MemStorage, url string, isCompress bool) error {
	var wg sync.WaitGroup
	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/gauge/%s/%v", url, metricName, metricValue)
			if err := sendRequest(f.ContentTypeTextPlain, isCompress, metricURL, nil); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/counter/%s/%v", url, metricName, metricValue)
			if err := sendRequest(f.ContentTypeTextPlain, isCompress, metricURL, nil); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
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
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	defer resp.Close()

	if !resp.Ok {
		return fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	return nil
}
