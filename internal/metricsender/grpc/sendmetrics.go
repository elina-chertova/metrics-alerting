package grpc

import (
	"context"
	pb "github.com/elina-chertova/metrics-alerting.git/api/proto"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"go.uber.org/zap"
	"time"
)

type SenderGRPC struct {
	Client pb.MetricsServiceClient
	Config *config.Agent
}

func (s *SenderGRPC) SendMetrics(storage *filememory.MemStorage) error {
	var metrics []*pb.Metric

	for metricName, metricValue := range storage.Gauge {
		metric := &pb.Metric{
			Id:    metricName,
			Type:  config.Gauge,
			Value: metricValue,
		}
		metrics = append(metrics, metric)
	}

	for metricName, metricValue := range storage.Counter {
		metric := &pb.Metric{
			Id:    metricName,
			Type:  config.Counter,
			Delta: metricValue,
		}
		metrics = append(metrics, metric)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.Client.UpdateBatchMetrics(ctx, &pb.UpdateBatchMetricsRequest{Metrics: metrics})
	if err != nil {
		zap.L().Error("could not update metrics", zap.Error(err))
		return err
	}
	return nil
}
