package asymencrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func EncryptDataWithPublicKey(data []byte, publicKeyPath string) ([]byte, error) {
	publicKeyFile, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	pubBlock, _ := pem.Decode(publicKeyFile)
	pubKeyValue, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		panic(err)
	}
	pub := pubKeyValue.(*rsa.PublicKey)
	encryptOAEP, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, data, nil)
	if err != nil {
		panic(err)
	}
	cipherByte := encryptOAEP
	return cipherByte, err
}
