package request

import (
	"fmt"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/levigross/grequests"
	"sync"
)

func sendRequest(name string, value any, metricsType string, url string) error {
	metricURL := fmt.Sprintf("%s/%s/%s/%v", url, metricsType, name, value)
	_, err := grequests.Post(metricURL, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return err
	}
	return nil
}

func MetricsToServer(s *st.MemStorage, url string) {
	var wg sync.WaitGroup
	for metricName, metricValue := range s.Gauge {
		wg.Add(1)
		go func(metricName string, metricValue any) {
			defer wg.Done()
			err := sendRequest(metricName, metricValue, "gauge", url)
			if err != nil {
				return
			}
		}(metricName, metricValue)
	}

	for metricName, metricValue := range s.Counter {
		wg.Add(1)
		go func(metricName string, metricValue any) {
			defer wg.Done()
			err := sendRequest(metricName, metricValue, "counter", url)
			if err != nil {
				return
			}
		}(metricName, metricValue)
	}
	wg.Wait()
}
