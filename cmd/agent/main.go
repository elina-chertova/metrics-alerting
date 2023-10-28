package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"time"
)

func main() {
	var urlUpdate string
	agentConfig := config.NewAgent()
	settings := config.NewSettings()
	storage := filememory.NewMemStorage(false, nil)

	if settings.IsSendBatch {
		urlUpdate = "http://" + agentConfig.FlagAddress + "/updates"
	} else {
		urlUpdate = "http://" + agentConfig.FlagAddress + "/update"
	}

	go func() {
		for {
			st.ExtractMetrics(storage)
			time.Sleep(time.Duration(agentConfig.PollInterval) * time.Second)
		}
	}()
	go func() {
		for {
			time.Sleep(time.Duration(agentConfig.ReportInterval) * time.Second)
			r.MetricsToServer(
				storage,
				settings.FlagContentType,
				urlUpdate,
				settings.IsCompress,
				settings.IsSendBatch,
				agentConfig.SecretKey,
			)
		}
	}()

	select {}
}
