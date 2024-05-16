package metricsender

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Worker struct {
	Settings *config.Settings
	Config   *config.Agent
	once     sync.Once
}

type MetricsSender interface {
	SendMetrics(
		s *filememory.MemStorage,
	) error
}

func SendMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
	sender MetricsSender,
	stopChan <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	worker.once.Do(
		func() {
			if worker.Settings.IsSendBatch {
				worker.Settings.URL = fmt.Sprintf(
					worker.Settings.URL,
					worker.Config.FlagAddress,
					"updates",
				)
			} else {
				worker.Settings.URL = fmt.Sprintf(
					worker.Settings.URL,
					worker.Config.FlagAddress,
					"update",
				)
			}
		},
	)

	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
			}

			time.Sleep(time.Duration(worker.Config.ReportInterval) * time.Second)

			err := sender.SendMetrics(
				storage,
			)

			if err != nil {
				logger.Log.Error(err.Error(), zap.String("method", "MetricsToServer"))
			}
		}
	}()
}
