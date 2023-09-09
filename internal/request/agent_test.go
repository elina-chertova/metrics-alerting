package request

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	name := "metricName"
	value := 42
	metricsType := "counter"

	err := sendRequest(name, value, metricsType, server.URL)
	assert.NoError(t, err)
}
