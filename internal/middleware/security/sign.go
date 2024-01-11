// Package security provides functionalities related to security aspects
// of the application.
package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Hash generates a HMAC SHA256 hash of the provided data using a secret key.
//
// Parameters:
// - data: The string data to be hashed.
// - secretKey: A byte slice representing the secret key used for hashing.
//
// Returns:
// - A string representing the hex-encoded hash of the data.
func Hash(data string, secretKey []byte) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	hashDst := h.Sum(nil)
	return hex.EncodeToString(hashDst)
}

var (
	ErrInvalidHash = errors.New("invalid hash")
	ErrReadReqBody = errors.New("error reading request body")
)

// CheckHash compares a correct hash with a request hash and returns an error
// if they do not match.
//
// Parameters:
// - correctHash: The correct hash to compare against.
// - requestHash: The hash obtained from the request.
//
// Returns:
// - An error if the hashes do not match, indicating a potential data integrity issue.
func CheckHash(correctHash, requestHash string) error {
	decodeRightHash, _ := hex.DecodeString(correctHash)
	decodeRequestHash, _ := hex.DecodeString(requestHash)

	if !bytes.Equal(decodeRightHash, decodeRequestHash) {
		return fmt.Errorf(ErrInvalidHash.Error())
	}
	return nil
}

// HashCheckMiddleware returns a Gin middleware function that validates the hash
// of the request body. It checks if the request hash matches the computed hash
// of the request body using a secret key.
//
// Parameters:
// - secretKey: A string representing the secret key used for hashing.
//
// Returns:
// - A gin.HandlerFunc that performs hash validation on incoming requests.
func HashCheckMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secretKey == "" {
			return
		}
		requestHash := c.Request.Header.Get("HashSHA256")
		if requestHash == "" {
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": ErrReadReqBody.Error()},
			)
			return
		}
		correctHash := Hash(string(body), []byte(secretKey))

		err = CheckHash(correctHash, requestHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrInvalidHash.Error()})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}
