package formatter

import (
	"encoding/json"
	"testing"
)

func TestMetrics_MarshalJSON(t *testing.T) {
	var v2 = 20.5
	metrics := Metric{
		ID:    "metric1",
		MType: "type1",
		Delta: nil,
		Value: &v2,
	}

	_, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Error marshalling Metrics to JSON: %v", err)
	}
}
