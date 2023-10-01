package formatter

import (
	"encoding/json"
	"testing"
)

func TestMetrics_MarshalJSON(t *testing.T) {
	var v1 int64 = 10
	var v2 = 20.5
	metrics := Metrics{
		metrics: []Metric{
			{
				ID:    "metric1",
				MType: "type1",
				Delta: nil,
				Value: &v2,
			},
			{
				ID:    "metric2",
				MType: "type2",
				Delta: &v1,
				Value: nil,
			},
		},
	}

	_, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Error marshalling Metrics to JSON: %v", err)
	}
}
