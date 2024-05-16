package grpc

import (
	"context"
	"fmt"
	pb "github.com/elina-chertova/metrics-alerting.git/api/proto"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	serviceInterface "github.com/elina-chertova/metrics-alerting.git/internal/handlers"
)

// Handler encapsulates handling logic for metric-related HTTP endpoints.
type Handler struct {
	memStorage serviceInterface.MetricsStorage
}

// NewHandler creates a new Handler with the given metrics storage.
func NewHandler(st serviceInterface.MetricsStorage) *Handler {
	return &Handler{st}
}

type Server struct {
	pb.UnimplementedMetricsServiceServer
	Handler   *Handler
	SecretKey string
	CryptoKey string
}

func (s *Server) UpdateBatchMetrics(
	ctx context.Context,
	req *pb.UpdateBatchMetricsRequest,
) (*pb.UpdateBatchMetricsResponse, error) {
	var metrics []f.Metric

	for _, m := range req.Metrics {
		var metric f.Metric

		switch m.Type {
		case config.Gauge:
			metric = f.Metric{
				ID:    m.Id,
				MType: config.Gauge,
				Value: &m.Value,
			}
		case config.Counter:
			metric = f.Metric{
				ID:    m.Id,
				MType: config.Counter,
				Delta: &m.Delta,
			}
		default:
			return &pb.UpdateBatchMetricsResponse{Status: "Failed"}, fmt.Errorf("unsupported metric type")
		}

		metrics = append(metrics, metric)
	}
	err := s.Handler.memStorage.InsertBatchMetrics(metrics)
	if err != nil {
		return &pb.UpdateBatchMetricsResponse{Status: "Failed"}, err
	}
	return &pb.UpdateBatchMetricsResponse{Status: "Success"}, nil
}
