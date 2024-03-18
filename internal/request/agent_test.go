package request

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendRequest(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	err := generateTestKeys()
	require.NoError(t, err)

	publicKeyPath := "publicKey.pem"
	privateKeyPath := "privateKey.pem"
	ip := getIP()

	err = sendRequest("application/json", false, server.URL, nil, "", publicKeyPath, ip)
	assert.NoError(t, err)
	os.Remove(publicKeyPath)
	os.Remove(privateKeyPath)
}

func generateTestKeys() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	privateFile, err := os.Create("privateKey.pem")
	if err != nil {
		return err
	}
	defer privateFile.Close()
	if err := pem.Encode(
		privateFile,
		&pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes},
	); err != nil {
		return err
	}

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	publicFile, err := os.Create("publicKey.pem")
	if err != nil {
		return err
	}
	defer publicFile.Close()
	if err := pem.Encode(
		publicFile,
		&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyBytes},
	); err != nil {
		return err
	}

	return nil
}

func getIP() net.IP {
	dial, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalf("Failed to get udp connection: %v", err)
	}
	defer dial.Close()
	localAddr := dial.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
