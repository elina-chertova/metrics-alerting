package main

import (
	"fmt"
	"sync"
	"time"

	_ "net/http/pprof"

	"go.uber.org/zap"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	st "github.com/elina-chertova/metrics-alerting.git/internal/metricextractor"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func sendMetricsWorker(storage *filememory.MemStorage, worker *Worker, stopChan <-chan struct{}) {
	worker.once.Do(
		func() {
			if worker.settings.IsSendBatch {
				worker.settings.URL = fmt.Sprintf(
					worker.settings.URL,
					worker.config.FlagAddress,
					"updates",
				)
			} else {
				worker.settings.URL = fmt.Sprintf(
					worker.settings.URL,
					worker.config.FlagAddress,
					"update",
				)
			}
		},
	)

	go func() {
		for {
			select {
			case <-stopChan:
			default:
			}

			time.Sleep(time.Duration(worker.config.ReportInterval) * time.Second)

			err := r.MetricsToServer(
				storage,
				worker.settings.FlagContentType,
				worker.settings.URL,
				worker.settings.IsCompress,
				worker.settings.IsSendBatch,
				worker.config.SecretKey,
				worker.config.CryptoKey,
			)
			if err != nil {
				logger.Log.Error(err.Error(), zap.String("method", "MetricsToServer"))
			}
			logger.Log.Info(
				"Metrics sent", zap.String("method", "MetricsToServer"),
				zap.String("URL", worker.settings.URL),
				zap.String("ContentType", worker.settings.FlagContentType),
				zap.Bool("Compressed", worker.settings.IsCompress),
				zap.Bool("Batch", worker.settings.IsSendBatch),
			)
		}
	}()
}

func extractMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
	stopChan <-chan struct{},
) {
	for {
		select {
		case <-stopChan:
		default:
		}
		err := st.ExtractMetrics(storage)
		if err != nil {
			logger.Log.Error(err.Error(), zap.String("method", "ExtractMetrics"))
		}
		time.Sleep(time.Duration(worker.config.PollInterval) * time.Second)
	}

}

func extractOSMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
	stopChan <-chan struct{},
) {
	for {
		select {
		case <-stopChan:
		default:
		}
		err := st.ExtractOSMetrics(storage)
		if err != nil {
			logger.Log.Error(err.Error(), zap.String("method", "ExtractOSMetrics"))
		}
		time.Sleep(time.Duration(worker.config.PollInterval) * time.Second)
	}

}

type Worker struct {
	settings *config.Settings
	config   *config.Agent
	once     sync.Once
}

func main() {
	fmt.Printf("Build version:%s\n", buildVersion)
	fmt.Printf("Build date:%s\n", buildDate)
	fmt.Printf("Build commit:%s\n", buildCommit)

	w := &Worker{
		settings: config.NewSettings(),
		config:   config.NewAgent(),
	}
	logger.LogInit("info")
	storage := filememory.NewMemStorage(false, nil)
	stopCh := make(chan struct{})
	var numSendWorkers = w.config.RateLimit
	var numExtractWorkers = w.config.RateLimit

	for sw := 0; sw < numSendWorkers; sw++ {
		go sendMetricsWorker(storage, w, stopCh)
	}
	for ew := 0; ew < numExtractWorkers; ew++ {
		go extractMetricsWorker(storage, w, stopCh)
		go extractOSMetricsWorker(storage, w, stopCh)
	}

	close(stopCh)
	select {}
}
