package metricextractor

import (
	sender "github.com/elina-chertova/metrics-alerting.git/internal/metricsender"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"go.uber.org/zap"
	"sync"
	"time"
)

func ExtractMetricsWorker(
	storage *filememory.MemStorage,
	worker *sender.Worker,
	stopChan <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-stopChan:
			return
		default:
		}
		err := ExtractMetrics(storage)
		if err != nil {
			logger.Log.Error(err.Error(), zap.String("method", "ExtractMetrics"))
		}
		time.Sleep(time.Duration(worker.Config.PollInterval) * time.Second)
	}

}

func ExtractOSMetricsWorker(
	storage *filememory.MemStorage,
	worker *sender.Worker,
	stopChan <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-stopChan:
			return
		default:
		}
		err := ExtractOSMetrics(storage)
		if err != nil {
			logger.Log.Error(err.Error(), zap.String("method", "ExtractOSMetrics"))
		}
		time.Sleep(time.Duration(worker.Config.PollInterval) * time.Second)
	}

}
