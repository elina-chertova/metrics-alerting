package request

import (
	"encoding/json"
	"fmt"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/levigross/grequests"
	"sync"
)

const (
	ContentTypeJSON      = "application/json"
	ContentTypeTextPlain = "text/plain"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func MetricsToServer(s *st.MemStorage, contentType string, url string) error {
	switch contentType {
	case ContentTypeTextPlain:
		return metricsToServerTextPlain(s, url)
	case ContentTypeJSON:
		return metricsToServerAppJSON(s, url)
	default:
		return fmt.Errorf("error creating HTTP request, wrong Content-Type: %s", contentType)
	}
}

func (m Metrics) MarshalJSON() ([]byte, error) {
	type MetricAlias Metric
	var jsonData []interface{}
	for _, metric := range m.metrics {
		if metric.Delta == nil {
			aliasValue := struct {
				MetricAlias
				Value float64 `json:"value"`
			}{
				MetricAlias: MetricAlias(metric),
				Value:       *metric.Value,
			}
			jsonData = append(jsonData, aliasValue)
		} else if metric.Value == nil {
			aliasValue := struct {
				MetricAlias
				Delta int64 `json:"delta"`
			}{
				MetricAlias: MetricAlias(metric),
				Delta:       *metric.Delta,
			}
			jsonData = append(jsonData, aliasValue)
		}
	}
	return json.Marshal(jsonData)
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type Metrics struct {
	metrics []Metric
}

func metricsToServerAppJSON(s *st.MemStorage, url string) error {
	var metrics []Metric
	for name, value := range s.Gauge {
		val := value
		metrics = append(
			metrics, Metric{
				ID:    name,
				MType: st.Gauge,
				Value: &val,
			},
		)
	}
	for name, value := range s.Counter {
		val := value
		metrics = append(
			metrics, Metric{
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
	return sendRequest(ContentTypeJSON, url, out)
}

func metricsToServerTextPlain(s *st.MemStorage, url string) error {
	var wg sync.WaitGroup
	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue float64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/gauge/%s/%v", url, metricName, metricValue)
			if err := sendRequest(ContentTypeTextPlain, metricURL, nil); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue int64) {
			defer wg.Done()
			metricURL := fmt.Sprintf("%s/counter/%s/%v", url, metricName, metricValue)
			if err := sendRequest(ContentTypeTextPlain, metricURL, nil); err != nil {
				fmt.Printf("Error sending request for %s: %v\n", metricName, err)
			}
		}(metricName, metricValue)
	}

	wg.Wait()
	return nil
}

func sendRequest(contentType string, url string, jsonBody []byte) error {
	ro := &grequests.RequestOptions{
		Headers: map[string]string{"Content-Type": contentType},
		JSON:    jsonBody,
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
