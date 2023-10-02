package store

import "github.com/elina-chertova/metrics-alerting.git/internal/storage/metrics"

type storager struct{ memStorage *metrics.MemStorage }

func NewStorager(s *metrics.MemStorage) *storager {
	return &storager{s}
}
