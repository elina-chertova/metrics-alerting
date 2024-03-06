// Package request contains functions for sending metric data to a server.
package request

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
)

// MetricsToServer sends metrics stored in memory to a specified server.
// It supports sending metrics in batch or individually, with optional compression,
// and handles different content types (JSON, TextPlain).
//
// Parameters:
// - s: The in-memory storage containing the metrics to be sent.
// - contentType: The MIME type of the content to be sent.
// - url: The URL of the server where the metrics are to be sent.
// - isCompress: Indicates whether the data should be compressed before sending.
// - isSendBatch: Determines whether to send metrics in batches.
// - secretKey: A secret key used for secure communication (optional).
//
// Returns:
// - An error if sending the metrics fails or if an invalid content type is specified.
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

// BackoffSendRequest attempts to send a request to a server with exponential backoff
// retry strategy.
//
// Parameters:
// - contentType: The MIME type of the content to be sent.
// - url: The URL of the server where the data is to be sent.
// - isCompress: Indicates whether the data should be compressed.
// - out: The byte slice of the data to be sent.
// - secretKey: A secret key used for secure communication (optional).
//
// Returns:
// - An error if all attempts to send the request fail.
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

// metricsToServerBatch sends a batch of all collected metrics to the specified server URL.
// It converts the metrics stored in memory into a JSON format and sends them in a single request.
//
// Parameters:
// - s: An instance of in-memory storage containing the metrics to be sent.
// - url: The server URL to which the metrics are to be sent.
// - isCompress: Indicates whether the data should be compressed before sending.
// - secretKey: A secret key used for secure communication (optional).
//
// Returns:
// - An error if there is an issue in creating JSON or sending the request.
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

// metricsToServerAppJSON sends individual metrics in JSON format to the server URL
// using concurrent goroutines. Each metric is sent as a separate request.
//
// Parameters:
// - s: An instance of in-memory storage containing the metrics to be sent.
// - url: The server URL to which the metrics are to be sent.
// - isCompress: Indicates whether the data should be compressed before sending.
// - secretKey: A secret key used for secure communication (optional).
//
// Returns:
// - An error if encountered during the processing or sending of any metric.
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

// metricsToServerTextPlain sends individual metrics in plain text format to the server URL
// using concurrent goroutines. Each metric is sent as a separate request with the metric value
// included in the URL.
//
// Parameters:
// - s: An instance of in-memory storage containing the metrics to be sent.
// - url: The server URL to which the metrics are to be sent.
// - isCompress: Indicates whether the data should be compressed before sending.
// - secretKey: A secret key used for secure communication (optional).
//
// Returns:
// - An error if encountered during the processing or sending of any metric.
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
