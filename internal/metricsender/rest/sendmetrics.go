package rest

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"go.uber.org/zap"
	"log"
	"net"
)

func getIP() net.IP {
	dial, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalf("Failed to get udp connection: %v", err)
	}
	defer dial.Close()
	localAddr := dial.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

type SenderRest struct {
	Settings *config.Settings
	Config   *config.Agent
}

func (rest *SenderRest) SendMetrics(storage *filememory.MemStorage) error {
	ip := getIP()

	err := r.MetricsToServer(
		storage,
		rest.Settings.FlagContentType,
		rest.Settings.URL,
		rest.Settings.IsCompress,
		rest.Settings.IsSendBatch,
		rest.Config.SecretKey,
		rest.Config.CryptoKey,
		ip,
	)
	if err != nil {
		logger.Log.Error(err.Error(), zap.String("method", "MetricsToServer"))
		return err
	}
	logger.Log.Info(
		"Metrics sent", zap.String("method", "MetricsToServer"),
		zap.String("URL", rest.Settings.URL),
		zap.String("ContentType", rest.Settings.FlagContentType),
		zap.Bool("Compressed", rest.Settings.IsCompress),
		zap.Bool("Batch", rest.Settings.IsSendBatch),
	)
	return nil
}
