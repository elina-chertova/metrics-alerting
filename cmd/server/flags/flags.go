package flags

import "flag"

var FlagAddress string

func ParseServerFlags() {
	flag.StringVar(&FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
