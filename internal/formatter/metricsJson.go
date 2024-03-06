// Package formatter provides utilities for formatting and marshaling
// metrics data.
package formatter

import (
	"encoding/json"
)

// ContentTypeJSON specifies the MIME type for JSON content.
const ContentTypeJSON = "application/json"

// ContentTypeTextPlain specifies the MIME type for plain text content.
const ContentTypeTextPlain = "text/plain"

// Metric represents a measurement or other quantifiable data point in an application.
// It includes an identifier, a type, and optional delta and value fields.
// The struct is designed to be marshaled into JSON, handling nil delta or value appropriately.
type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// MarshalJSON customizes the JSON marshaling for Metric. It ensures that
// either 'delta' or 'value' is included in the JSON output depending on which
// is non-nil.
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

// Marshaler is an interface representing the ability to marshal an object into JSON.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
