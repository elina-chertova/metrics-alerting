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
	DatabaseDSN     string
}

func ParseServerFlags(s *Server) {
	flag.StringVar(&s.FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&s.StoreInterval, "i", 300, "seconds to save filememory data to server")
	flag.StringVar(
		&s.FileStoragePath,
		"f",
		"tmp/metrics-db.json",
		"temp file to save metrics",
	)
	flag.BoolVar(&s.FlagRestore, "r", true, "is load saved filememory during server start")
	flag.StringVar(
		&s.DatabaseDSN,
		"d",
		"",
		"Database DSN. Ex: postgres://postgres:123qwe@localhost:5432/metrics_db",
	)

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
	if envRunDSN := os.Getenv("DATABASE_DSN"); envRunDSN != "" {
		s.DatabaseDSN = envRunDSN
	}

}

func NewServer() *Server {
	s := &Server{}
	ParseServerFlags(s)
	return s
}
