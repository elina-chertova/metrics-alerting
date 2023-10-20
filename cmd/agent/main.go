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
	storage := filememory.NewMemStorage(false, nil)

	flagContentType := "application/json"
	isCompress := true
	isSendBatch := false
	if isSendBatch {
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
			r.MetricsToServer(storage, flagContentType, urlUpdate, isCompress, isSendBatch)
		}
	}()

	select {}
}
