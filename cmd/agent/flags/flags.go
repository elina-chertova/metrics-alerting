package flags

import (
	"flag"
)

var FlagAddress string
var PollInterval int64
var ReportInterval int64

func ParseAgentFlags() {
	flag.StringVar(&FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&PollInterval, "p", 2, "time in seconds to update metrics, example: 2")
	flag.Int64Var(&ReportInterval, "r", 10, "time in seconds to send data to server, example: 10")
	flag.Parse()
}
