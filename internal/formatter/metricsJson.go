package formatter

import (
	"encoding/json"
)

const (
	ContentTypeJSON      = "application/json"
	ContentTypeTextPlain = "text/plain"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m Metrics) MarshalJSON() ([]byte, error) {
	type MetricAlias Metric
	var jsonData []interface{}
	for _, metric := range m.metrics {
		if metric.Delta == nil {
			aliasValue := struct {
				MetricAlias
				Value float64 `json:"value"`
			}{
				MetricAlias: MetricAlias(metric),
				Value:       *metric.Value,
			}
			jsonData = append(jsonData, aliasValue)
		} else if metric.Value == nil {
			aliasValue := struct {
				MetricAlias
				Delta int64 `json:"delta"`
			}{
				MetricAlias: MetricAlias(metric),
				Delta:       *metric.Delta,
			}
			jsonData = append(jsonData, aliasValue)
		}
	}
	return json.Marshal(jsonData)
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type Metrics struct {
	metrics []Metric
}
