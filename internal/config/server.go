package config

import (
	"flag"
	"os"
	"strconv"
)

type Server struct {
	FlagAddress     string
	StoreInterval   int
	FileStoragePath string
	FlagRestore     bool
}

func ParseServerFlags(s *Server) {
	flag.StringVar(&s.FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&s.StoreInterval, "i", 15, "seconds to save metrics data to server")
	flag.StringVar(
		&s.FileStoragePath,
		"f",
		"tmp/metrics-db.json",
		"temp file to save metrics",
	)
	flag.BoolVar(&s.FlagRestore, "r", true, "is load saved metrics during server start")
	flag.Parse()
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.FlagAddress = envRunAddr
	}
	if envRunStoreInterval := os.Getenv("STORE_INTERVAL"); envRunStoreInterval != "" {
		s.StoreInterval, _ = strconv.Atoi(envRunStoreInterval)
	}
	if envRunStoragePath := os.Getenv("FILE_STORAGE_PATH"); envRunStoragePath != "" {
		s.FileStoragePath = envRunStoragePath
	}
	if envRunRestore := os.Getenv("RESTORE"); envRunRestore != "" {
		s.FlagRestore, _ = strconv.ParseBool(envRunRestore)
	}

}

func NewServer() *Server {
	s := &Server{}
	ParseServerFlags(s)
	return s
}
