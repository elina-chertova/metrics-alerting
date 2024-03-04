package metricsender

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
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

func SendMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
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

			err := r.MetricsToServer(
				storage,
				worker.Settings.FlagContentType,
				worker.Settings.URL,
				worker.Settings.IsCompress,
				worker.Settings.IsSendBatch,
				worker.Config.SecretKey,
				worker.Config.CryptoKey,
			)
			if err != nil {
				logger.Log.Error(err.Error(), zap.String("method", "MetricsToServer"))
			}
			logger.Log.Info(
				"Metrics sent", zap.String("method", "MetricsToServer"),
				zap.String("URL", worker.Settings.URL),
				zap.String("ContentType", worker.Settings.FlagContentType),
				zap.Bool("Compressed", worker.Settings.IsCompress),
				zap.Bool("Batch", worker.Settings.IsSendBatch),
			)
		}
	}()
}
