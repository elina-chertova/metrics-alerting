package main

import (
	"fmt"
	"github.com/elina-chertova/metrics-alerting.git/internal/asymencrypt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "net/http/pprof"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	st "github.com/elina-chertova/metrics-alerting.git/internal/metricextractor"
	sender "github.com/elina-chertova/metrics-alerting.git/internal/metricsender"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
)

var (
	wg           sync.WaitGroup
	stopCh       = make(chan struct{})
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:%s\n", buildVersion)
	fmt.Printf("Build date:%s\n", buildDate)
	fmt.Printf("Build commit:%s\n", buildCommit)

	w := &sender.Worker{
		Settings: config.NewSettings(),
		Config:   config.NewAgent(),
	}

	cryptoKeyPing(w.Config.CryptoKey)

	logger.LogInit("info")
	storage := filememory.NewMemStorage(false, nil)

	startWorkers(storage, w, stopCh, &wg)

	waitForSignals(stopCh)

	wg.Wait()
}

func startWorkers(
	storage *filememory.MemStorage,
	w *sender.Worker,
	stopCh <-chan struct{},
	wg *sync.WaitGroup,
) {
	numSendWorkers := w.Config.RateLimit
	numExtractWorkers := w.Config.RateLimit

	wg.Add(numSendWorkers + 2*numExtractWorkers)

	for sw := 0; sw < numSendWorkers; sw++ {
		go sender.SendMetricsWorker(storage, w, stopCh, wg)
	}
	for ew := 0; ew < numExtractWorkers; ew++ {
		go st.ExtractMetricsWorker(storage, w, stopCh, wg)
		go st.ExtractOSMetricsWorker(storage, w, stopCh, wg)
	}
}

func initiateGracefulShutdown() {
	close(stopCh)
	wg.Wait()
}

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

func waitForSignals(stopCh chan<- struct{}) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigint
	close(stopCh)
}
