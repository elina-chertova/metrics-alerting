package config

import (
	"flag"
	"github.com/goccy/go-json"
	"log"
	"os"
	"strconv"
	"time"
)

type Agent struct {
	FlagAddress    string `json:"address"`
	PollInterval   int    `json:"poll_interval"`
	ReportInterval int    `json:"report_interval"`
	SecretKey      string
	RateLimit      int
	CryptoKey      string `json:"crypto_key"`
}

type AgentConfigJSON struct {
	Address        string `json:"address"`
	PollInterval   string `json:"poll_interval"`
	ReportInterval string `json:"report_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func ParseAgentFlags(a *Agent) {
	flag.StringVar(&a.FlagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&a.PollInterval, "p", 2, "time in seconds to update filememory, example: 2")
	flag.IntVar(&a.ReportInterval, "r", 10, "time in seconds to send data to server, example: 10")
	flag.StringVar(&a.SecretKey, "k", "", "secret key for hash")
	flag.IntVar(&a.RateLimit, "l", 2, "Rate limit to max workers number")
	flag.StringVar(
		&a.CryptoKey,
		"crypto-key",
		"",
		"crypto key public",
	)

	configFilePath := flag.String(
		"c",
		"",
		"path to config file",
	)
	configFilePathAlt := flag.String("config", "", "path to config file (alternative)")
	flag.Parse()
	err := readFromJSON(a, configFilePath, configFilePathAlt)
	if err != nil {
		log.Println("Error reading JSON: ", err)
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		a.FlagAddress = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		a.ReportInterval, _ = strconv.Atoi(envReportInterval)
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		a.PollInterval, _ = strconv.Atoi(envPollInterval)
	}
	if envRunKey := os.Getenv("KEY"); envRunKey != "" {
		a.SecretKey = envRunKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		a.RateLimit, _ = strconv.Atoi(envRateLimit)
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		a.CryptoKey = envCryptoKey
	}

}

func readFromJSON(a *Agent, configFilePath *string, configFilePathAlt *string) error {
	finalConfigPath := *configFilePath
	if *configFilePathAlt != "" {
		finalConfigPath = *configFilePathAlt
	}
	if finalConfigPath != "" {
		file, err := os.ReadFile(finalConfigPath)
		if err != nil {
			return err
		}
		var jsonConfig AgentConfigJSON
		if err := json.Unmarshal(file, &jsonConfig); err == nil {
			if flag.Lookup("a").Value.String() == "localhost:8080" {
				a.FlagAddress = jsonConfig.Address
			}

			if flag.CommandLine.Lookup("p").Value.String() == strconv.Itoa(2) && jsonConfig.PollInterval != "" {
				if dur, err := time.ParseDuration(jsonConfig.PollInterval); err == nil {
					a.PollInterval = int(dur.Seconds())
				} else {
					return err
				}
			}

			if flag.CommandLine.Lookup("r").Value.String() == strconv.Itoa(10) && jsonConfig.ReportInterval != "" {
				if dur, err := time.ParseDuration(jsonConfig.ReportInterval); err == nil {
					a.ReportInterval = int(dur.Seconds())
				} else {
					return err
				}
			}

			if flag.Lookup("crypto-key").Value.String() == "" {
				a.CryptoKey = jsonConfig.CryptoKey
			}
		}
	}
	return nil
}

func NewAgent() *Agent {
	a := &Agent{}
	ParseAgentFlags(a)
	return a
}
