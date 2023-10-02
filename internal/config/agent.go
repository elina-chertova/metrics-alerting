package config

import (
	"flag"
	"os"
	"strconv"
)

type Agent struct {
	FlagAddress    string
	PollInterval   int
	ReportInterval int
}

func ParseAgentFlags(a *Agent) {
	flag.StringVar(&a.FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&a.PollInterval, "p", 2, "time in seconds to update metrics, example: 2")
	flag.IntVar(&a.ReportInterval, "r", 10, "time in seconds to send data to server, example: 10")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		a.FlagAddress = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		a.ReportInterval, _ = strconv.Atoi(envReportInterval)
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		a.PollInterval, _ = strconv.Atoi(envPollInterval)
	}

}

func NewAgent() *Agent {
	a := &Agent{}
	ParseAgentFlags(a)
	return a
}
