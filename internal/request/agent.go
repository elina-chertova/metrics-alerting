package request

import (
	"encoding/json"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	metrics2 "github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"sync"
	"time"
)

func MetricsToServer(
	s *metrics2.MemStorage,
	contentType string,
	url string,
	isCompress,
	isSendBatch bool,
) error {
	if isSendBatch {
		return metricsToServerBatch(s, url, isCompress)
	}

	switch contentType {
	case f.ContentTypeTextPlain:
		return metricsToServerTextPlain(s, url, isCompress)
	case f.ContentTypeJSON:
		return metricsToServerAppJSON(s, url, isCompress)
	default:
		return fmt.Errorf("error creating HTTP request, wrong Content-Type: %s", contentType)
	}
}

func BackoffSendRequest(contentType, url string, isCompress bool, out []byte) error {
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxRetries := 3

	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			if retry-1 < len(retryDelays) {
				time.Sleep(retryDelays[retry-1])
			}
		}

		err := sendRequest(contentType, isCompress, url, out)

		if err != nil {
			if retry == maxRetries {
				return fmt.Errorf("error sending request: %v", err)
			}
			fmt.Printf("Error sending request (retry %d/%d): %v\n", retry+1, maxRetries, err)
		} else {
			break
		}
	}
	return nil
}

func metricsToServerBatch(s *metrics2.MemStorage, url string, isCompress bool) error {
	var metric f.Metric
	var metrics []f.Metric

	s.LockGauge()
	defer s.UnlockGauge()
	for metricName, metricValue := range s.Gauge {
		metric, _ = formJSON(metricName, metricValue, config.Gauge)
		metrics = append(metrics, metric)
	}

	s.LockCounter()
	defer s.UnlockCounter()
	for metricName, metricValue := range s.Counter {
		metric, _ = formJSON(metricName, metricValue, config.Counter)
		metrics = append(metrics, metric)
	}
	out, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	err = BackoffSendRequest(f.ContentTypeJSON, url, isCompress, out)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	return nil
}

func metricsToServerAppJSON(s *metrics2.MemStorage, url string, isCompress bool) error {
	var wg sync.WaitGroup

	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metrics, _ := formJSON(metricName, metricValue, config.Gauge)
			out, err := json.Marshal(metrics)
			if err != nil {
				fmt.Printf("error creating JSON: %v\n", err)
			}

			if err = BackoffSendRequest(f.ContentTypeJSON, url, isCompress, out); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metrics, _ := formJSON(metricName, metricValue, config.Counter)
			out, err := json.Marshal(metrics)

			if err != nil {
				fmt.Printf("error creating JSON: %v\n", err)
			}
			if err = BackoffSendRequest(f.ContentTypeJSON, url, isCompress, out); err != nil {
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
			if err := BackoffSendRequest(
				f.ContentTypeTextPlain,
				metricURL,
				isCompress,
				nil,
			); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/counter/%s/%v", url, metricName, metricValue)
			if err := BackoffSendRequest(
				f.ContentTypeTextPlain,
				metricURL,
				isCompress,
				nil,
			); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
}
