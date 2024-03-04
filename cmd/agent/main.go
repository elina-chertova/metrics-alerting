package main

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/asymencrypt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	wg     sync.WaitGroup
	stopCh = make(chan struct{})
)

func cryptoKeyPing(cryptoKeyPath string) {
	originalText := "test ping"
	_, err := asymencrypt.EncryptDataWithPublicKey(
		[]byte(originalText),
		cryptoKeyPath,
	)
	if err != nil {
		log.Printf("Failed to encrypt data: %v\n", err)
		initiateGracefulShutdown()
		return
	}
}

func main() {
	fmt.Printf("Build version:%s\n", buildVersion)
	fmt.Printf("Build date:%s\n", buildDate)
	fmt.Printf("Build commit:%s\n", buildCommit)

	w := &Worker{
		settings: config.NewSettings(),
		config:   config.NewAgent(),
	}

	cryptoKeyPing(w.config.CryptoKey)

	logger.LogInit("info")
	storage := filememory.NewMemStorage(false, nil)

	startWorkers(storage, w, stopCh, &wg)

	waitForSignals(stopCh)

	wg.Wait()
}
func initiateGracefulShutdown() {
	close(stopCh)
	wg.Wait()
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func sendMetricsWorker(
	storage *filememory.MemStorage,
	worker *Worker,
	stopChan <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
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
				return
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
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-stopChan:
			return
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
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-stopChan:
			return
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

func startWorkers(
	storage *filememory.MemStorage,
	w *Worker,
	stopCh <-chan struct{},
	wg *sync.WaitGroup,
) {
	numSendWorkers := w.config.RateLimit
	numExtractWorkers := w.config.RateLimit

	wg.Add(numSendWorkers + 2*numExtractWorkers)

	for sw := 0; sw < numSendWorkers; sw++ {
		go sendMetricsWorker(storage, w, stopCh, wg)
	}
	for ew := 0; ew < numExtractWorkers; ew++ {
		go extractMetricsWorker(storage, w, stopCh, wg)
		go extractOSMetricsWorker(storage, w, stopCh, wg)
	}
}

func waitForSignals(stopCh chan<- struct{}) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigint
	close(stopCh)
}
