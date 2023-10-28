package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

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

func CheckHash(correctHash, requestHash string) error {
	decodeRightHash, _ := hex.DecodeString(correctHash)
	decodeRequestHash, _ := hex.DecodeString(requestHash)

	if !bytes.Equal(decodeRightHash, decodeRequestHash) {
		return fmt.Errorf(ErrInvalidHash.Error())
	}
	return nil
}

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
