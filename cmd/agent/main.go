package main

import (
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	r "github.com/elina-chertova/metrics-alerting.git/internal/request"
	st "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	"time"
)

func main() {
	agentConfig := config.NewAgent()
	storage := st.NewMemStorage()
	urlUpdate := "http://" + agentConfig.FlagAddress + "/update"
	flagContentType := "application/json"
	//agentConfig.FlagContentType
	go func() {
		for {
			st.ExtractMetrics(storage)
			time.Sleep(time.Duration(agentConfig.PollInterval) * time.Second)
		}
	}()
	go func() {
		for {
			time.Sleep(time.Duration(agentConfig.ReportInterval) * time.Second)
			r.MetricsToServer(storage, flagContentType, urlUpdate)
		}
	}()

	select {}
}
