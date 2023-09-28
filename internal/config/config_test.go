package config

import (
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	prevAddressEnv := os.Getenv("ADDRESS")

	os.Setenv("ADDRESS", "test_server_address:1234")

	expectedServer := &Server{
		FlagAddress: "test_server_address:1234",
	}

	server := NewServer()

	if server.FlagAddress != expectedServer.FlagAddress {
		t.Errorf("FlagAddress is not set correctly")
	}

	os.Setenv("ADDRESS", prevAddressEnv)
}
