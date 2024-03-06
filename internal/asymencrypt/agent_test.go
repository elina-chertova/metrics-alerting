package asymencrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

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

func TestEncryptAndDecrypt(t *testing.T) {
	err := generateTestKeys()
	require.NoError(t, err)

	originalText := "Hello, RSA encryption and decryption!"
	publicKeyPath := "publicKey.pem"
	privateKeyPath := "privateKey.pem"

	encryptedText, err := EncryptDataWithPublicKey([]byte(originalText), publicKeyPath)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedText)

	decryptedText, err := DecryptDataWithPrivateKey(encryptedText, privateKeyPath)
	require.NoError(t, err)
	require.Equal(t, originalText, string(decryptedText))

	os.Remove(publicKeyPath)
	os.Remove(privateKeyPath)
}
