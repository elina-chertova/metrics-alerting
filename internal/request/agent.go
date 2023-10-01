package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/levigross/grequests"
	"sync"
)

func MetricsToServer(s *st.MemStorage, contentType string, url string) error {
	switch contentType {
	case f.ContentTypeTextPlain:
		return metricsToServerTextPlain(s, url)
	case f.ContentTypeJSON:
		return metricsToServerAppJSON(s, url)
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

func metricsToServerAppJSON(s *st.MemStorage, url string) error {
	isCompress := true
	var metrics []f.Metric
	for name, value := range s.Gauge {
		val := value
		metrics = append(
			metrics, f.Metric{
				ID:    name,
				MType: st.Gauge,
				Value: &val,
			},
		)
	}
	for name, value := range s.Counter {
		val := value
		metrics = append(
			metrics, f.Metric{
				ID:    name,
				MType: st.Counter,
				Delta: &val,
			},
		)
	}
	out, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}
	return sendRequest(f.ContentTypeJSON, isCompress, url, out)
}

func metricsToServerTextPlain(s *st.MemStorage, url string) error {
	var wg sync.WaitGroup
	isCompress := false
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
