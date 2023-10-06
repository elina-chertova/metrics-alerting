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

func (m Metric) MarshalJSON() ([]byte, error) {
	type MetricAlias Metric
	var aliasValue interface{}

	if m.Delta == nil {
		aliasValue = struct {
			MetricAlias
			Value float64 `json:"value"`
		}{
			MetricAlias: MetricAlias(m),
			Value:       *m.Value,
		}
	} else if m.Value == nil {
		aliasValue = struct {
			MetricAlias
			Delta int64 `json:"delta"`
		}{
			MetricAlias: MetricAlias(m),
			Delta:       *m.Delta,
		}

	}
	return json.Marshal(aliasValue)
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
