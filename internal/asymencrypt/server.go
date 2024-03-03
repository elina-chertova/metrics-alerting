package asymencrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

func DecryptDataWithPrivateKey(encryptedData []byte, privateKeyPath string) ([]byte, error) {
	privateKeyFile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyFile)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing the key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	decryptedData, err := rsa.DecryptOAEP(
		sha1.New(),
		rand.Reader,
		priv,
		encryptedData,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
