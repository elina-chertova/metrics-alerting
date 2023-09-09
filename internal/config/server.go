package config

import (
	"flag"
	"os"
)

type Server struct {
	FlagAddress string
}

func ParseServerFlags(s *Server) {
	flag.StringVar(&s.FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.FlagAddress = envRunAddr
	}
}

func NewServer() *Server {
	s := &Server{}
	ParseServerFlags(s)
	return s
}
