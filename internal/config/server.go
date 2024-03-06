package config

import (
	"flag"
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"strconv"
	"time"
)

type Server struct {
	FlagAddress     string `json:"address"`
	StoreInterval   int    `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	FlagRestore     bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	SecretKey       string
	CryptoKey       string `json:"crypto_key"`
}

type ServerConfigJSON struct {
	Address         string `json:"address"`
	StoreInterval   string `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	CryptoKey       string `json:"crypto_key"`
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
	flag.StringVar(&s.SecretKey, "k", "", "secret key for hash")
	flag.StringVar(
		&s.CryptoKey,
		"crypto-key",
		"",
		"crypto key private",
	)

	configFilePath := flag.String(
		"c",
		"",
		"path to config file",
	)
	configFilePathAlt := flag.String("config", "", "path to config file (alternative)")
	flag.Parse()

	finalConfigPath := *configFilePath
	if *configFilePathAlt != "" {
		finalConfigPath = *configFilePathAlt
	}

	if finalConfigPath != "" {
		file, err := os.ReadFile(finalConfigPath)
		if err != nil {
			panic(err)
		}

		var jsonConfig ServerConfigJSON
		if err := json.Unmarshal(file, &jsonConfig); err != nil {
			panic(err)
		}

		if flag.Lookup("a").Value.String() == "localhost:8080" {
			s.FlagAddress = jsonConfig.Address
		}

		if flag.CommandLine.Lookup("i").Value.String() == strconv.Itoa(300) && jsonConfig.StoreInterval != "" {
			if dur, err := time.ParseDuration(jsonConfig.StoreInterval); err == nil {
				s.StoreInterval = int(dur.Seconds())
				fmt.Println("s.StoreInterval", s.StoreInterval)
			} else {
				panic(err)
			}
		}
		if flag.Lookup("f").Value.String() == "tmp/metrics-db.json" {
			s.FileStoragePath = jsonConfig.FileStoragePath
		}
		if !flag.Lookup("r").Value.(flag.Getter).Get().(bool) {
			s.FlagRestore = jsonConfig.Restore
			fmt.Println("s.FlagRestore", s.FlagRestore)
		}
		if flag.Lookup("d").Value.String() == "" {
			s.DatabaseDSN = jsonConfig.DatabaseDSN
			fmt.Println("s.DatabaseDSN", s.DatabaseDSN)
		}
		if flag.Lookup("crypto-key").Value.String() == "" {
			s.CryptoKey = jsonConfig.CryptoKey
		}
	}

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
	if envRunKey := os.Getenv("KEY"); envRunKey != "" {
		s.SecretKey = envRunKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		s.CryptoKey = envCryptoKey
	}

}

func NewServer() *Server {
	s := &Server{}
	ParseServerFlags(s)
	return s
}
