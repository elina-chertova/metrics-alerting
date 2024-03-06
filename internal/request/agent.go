package request

import (
	"encoding/json"
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"sync"
	"time"
)

func MetricsToServer(
	s *filememory.MemStorage,
	contentType string,
	url string,
	isCompress,
	isSendBatch bool,
	secretKey string,
) error {
	if isSendBatch {
		return metricsToServerBatch(s, url, isCompress, secretKey)
	}

	switch contentType {
	case formatter.ContentTypeTextPlain:
		return metricsToServerTextPlain(s, url, isCompress, secretKey)
	case formatter.ContentTypeJSON:
		return metricsToServerAppJSON(s, url, isCompress, secretKey)
	default:
		return fmt.Errorf("error creating HTTP request, wrong Content-Type: %s", contentType)
	}
}

func BackoffSendRequest(
	contentType, url string,
	isCompress bool,
	out []byte,
	secretKey string,
) error {
	retryDelays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxRetries := 3

	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			if retry-1 < len(retryDelays) {
				time.Sleep(retryDelays[retry-1])
			}
		}

		err := sendRequest(contentType, isCompress, url, out, secretKey)

		if err != nil {
			if retry == maxRetries {
				return fmt.Errorf("error sending request: %v", err)
			}
			logger.Log.Error(
				fmt.Sprintf(
					"Error sending request (retry %d/%d): %v",
					retry+1,
					maxRetries,
					err,
				),
			)
		} else {
			break
		}
	}
	return nil
}

func metricsToServerBatch(
	s *filememory.MemStorage,
	url string,
	isCompress bool,
	secretKey string,
) error {
	var metric formatter.Metric
	var metrics []formatter.Metric

	for metricName, metricValue := range s.Gauge {
		metric, _ = formJSON(metricName, metricValue, config.Gauge)
		metrics = append(metrics, metric)
	}

	for metricName, metricValue := range s.Counter {
		metric, _ = formJSON(metricName, metricValue, config.Counter)
		metrics = append(metrics, metric)
	}
	out, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error creating JSON: %v", err)
	}

	err = BackoffSendRequest(formatter.ContentTypeJSON, url, isCompress, out, secretKey)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	return nil
}

func metricsToServerAppJSON(
	s *filememory.MemStorage,
	url string,
	isCompress bool,
	secretKey string,
) error {
	var wg sync.WaitGroup

	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metrics, _ := formJSON(metricName, metricValue, config.Gauge)
			out, err := json.Marshal(metrics)
			if err != nil {
				logger.Log.Error(fmt.Sprintf("error creating JSON: %v", err))
			}

			if err = BackoffSendRequest(
				formatter.ContentTypeJSON,
				url,
				isCompress,
				out,
				secretKey,
			); err != nil {
				logger.Log.Error(fmt.Sprintf("Error sending request for %s: %v", metricName, err))
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
				logger.Log.Error(fmt.Sprintf("error creating JSON: %v", err))
			}
			if err = BackoffSendRequest(
				formatter.ContentTypeJSON,
				url,
				isCompress,
				out,
				secretKey,
			); err != nil {
				logger.Log.Error(fmt.Sprintf("Error sending request for %s: %v", metricName, err))
			}

		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
}

func metricsToServerTextPlain(
	s *filememory.MemStorage,
	url string,
	isCompress bool,
	secretKey string,
) error {
	var wg sync.WaitGroup

	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/gauge/%s/%v", url, metricName, metricValue)
			if err := BackoffSendRequest(
				formatter.ContentTypeTextPlain,
				metricURL,
				isCompress,
				nil,
				secretKey,
			); err != nil {
				logger.Log.Error(fmt.Sprintf("Error sending request for %s: %v", metricName, err))
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/counter/%s/%v", url, metricName, metricValue)
			if err := BackoffSendRequest(
				formatter.ContentTypeTextPlain,
				metricURL,
				isCompress,
				nil,
				secretKey,
			); err != nil {
				logger.Log.Error(fmt.Sprintf("Error sending request for %s: %v", metricName, err))
			}
		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
}
