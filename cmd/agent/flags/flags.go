package flags

import (
	"flag"
	"time"
)

var FlagAddress string
var PollInterval time.Duration
var ReportInterval time.Duration

func ParseAgentFlags() {
	flag.StringVar(&FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.DurationVar(&PollInterval, "p", 2*time.Second, "time in seconds to update metrics, example: 2s")
	flag.DurationVar(&ReportInterval, "r", 10*time.Second, "time in seconds to send data to server, example: 10s")
	flag.Parse()
}
