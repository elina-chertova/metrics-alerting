package main

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"sync"
	"time"
)

func sendMetricsWorker(storage *filememory.MemStorage, worker *Worker, stopChan <-chan struct{}) {

	worker.once.Do(
		func() {
			if worker.settings.IsSendBatch {
				worker.settings.Url = fmt.Sprintf(
					worker.settings.Url,
					worker.config.FlagAddress,
					"updates",
				)
			} else {
				worker.settings.Url = fmt.Sprintf(
					worker.settings.Url,
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

			r.MetricsToServer(
				storage,
				worker.settings.FlagContentType,
				worker.settings.Url,
				worker.settings.IsCompress,
				worker.settings.IsSendBatch,
				worker.config.SecretKey,
			)
		}
	}()
}

func extractMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
	stopChan <-chan struct{},
) {
	go func() {
		for {
			select {
			case <-stopChan:
			default:
			}
			st.ExtractMetrics(storage)
			time.Sleep(time.Duration(worker.config.PollInterval) * time.Second)
		}
	}()
}

type Worker struct {
	settings *config.Settings
	config   *config.Agent
	once     sync.Once
}

func main() {
	w := &Worker{
		settings: config.NewSettings(),
		config:   config.NewAgent(),
	}

	storage := filememory.NewMemStorage(false, nil)
	stopCh := make(chan struct{})
	var numSendWorkers = 2
	var numExtractWorkers = 2

	for sw := 0; sw < numSendWorkers; sw++ {
		go sendMetricsWorker(storage, w, stopCh)
	}
	for ew := 0; ew < numExtractWorkers; ew++ {
		go extractMetricsWorker(storage, w, stopCh)
	}

	close(stopCh)
	select {}
}
