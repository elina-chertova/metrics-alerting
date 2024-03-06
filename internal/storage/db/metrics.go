// Package db provides database operations for managing metrics data.
// It includes functions to update, retrieve, and insert metrics into a database,
// supporting both gauge and counter types. This package uses the GORM library
// for database interactions, offering a high-level abstraction for database operations.
package db

import (
	"errors"
	"fmt"

	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	"github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"gorm.io/gorm"
)

var (
	ErrRetrieveMetric    = errors.New("failed to retrieve metric")
	ErrSaveMetric        = errors.New("failed to save metric")
	ErrCreateMetric      = errors.New("failed to create metric")
	ErrCommitTransaction = errors.New("transaction commit error")
)

// TypeIsCounter is a GORM scope function that filters database queries to return only counter metrics.
func TypeIsCounter(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", config.Counter)
}

// TypeIsGauge is a GORM scope function that filters database queries to return only gauge metrics.
func TypeIsGauge(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", config.Gauge)
}

// UpdateCounter updates the value of a counter metric in the database.
//
// Parameters:
// - name: The name of the counter metric.
// - value: The value to be added to the counter.
// - ok: A boolean indicating if the metric already exists in the database.
//
// Returns:
// - An error if the update or creation of the metric fails.
func (db DB) UpdateCounter(name string, value int64, ok bool) error {
	var m Metrics
	if ok {
		result := db.Database.Scopes(TypeIsCounter).Where("name = ?", name).Order("").First(&m)
		if result.Error != nil {
			logger.Log.Error(fmt.Sprintf("%s: %v", ErrRetrieveMetric, result.Error))
		}

		m.Delta += value

		result = db.Database.Save(&m)

		if result.Error != nil {
			return fmt.Errorf("%s: %v", ErrSaveMetric, result.Error)
		}

		return nil
	}
	data := db.Database.Create(
		&Metrics{
			Name:  name,
			Type:  config.Counter,
			Delta: value,
		},
	)
	if data.Error != nil {
		return fmt.Errorf("%s: %v", ErrCreateMetric, data.Error)
	}
	return nil
}

// UpdateGauge updates the value of a gauge metric in the database.
//
// Parameters:
// - name: The name of the gauge metric.
// - value: The value to be set for the gauge metric.
//
// Returns:
// - An error if the update or creation of the metric fails.
func (db DB) UpdateGauge(name string, value float64) error {
	var m Metrics

	if result := db.Database.Scopes(TypeIsGauge).Where(
		"name = ?",
		name,
	).Order("").First(&m); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		data := db.Database.Create(
			&Metrics{
				Name:  name,
				Type:  config.Gauge,
				Value: value,
			},
		)
		if data.Error != nil {
			return fmt.Errorf("%s: %v", ErrCreateMetric, data.Error)
		}
		return nil
	}
	m.Value = value

	result := db.Database.Save(&m)
	if result.Error != nil {
		return fmt.Errorf("%s: %v", ErrSaveMetric, result.Error)
	}
	return nil
}

// GetCounter retrieves the value of a counter metric from the database.
//
// Parameters:
// - name: The name of the counter metric to retrieve.
//
// Returns:
// - The value of the counter metric.
// - A boolean indicating if the metric was found in the database.
// - An error if the retrieval fails.
func (db DB) GetCounter(name string) (int64, bool, error) {
	var m Metrics
	result := db.Database.Scopes(TypeIsCounter).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	return m.Delta, true, nil
}

// GetGauge retrieves the value of a gauge metric from the database.
//
// Parameters:
// - name: The name of the gauge metric to retrieve.
//
// Returns:
// - The value of the gauge metric.
// - A boolean indicating if the metric was found in the database.
// - An error if the retrieval fails.
func (db DB) GetGauge(name string) (float64, bool, error) {
	var m Metrics
	result := db.Database.Scopes(TypeIsGauge).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	return m.Value, true, nil
}

// GetMetrics retrieves all counter and gauge metrics from the database.
//
// Returns:
// - A map of counter metrics with their names and values.
// - A map of gauge metrics with their names and values.
func (db DB) GetMetrics() (map[string]int64, map[string]float64) {
	var counterStruct []struct {
		Name  string
		Delta int64
	}
	var gaugeStruct []struct {
		Name  string
		Value float64
	}

	var m filememory.MemStorage
	db.Database.Table("metrics").Select("name, delta").Scopes(TypeIsCounter).Order("").Scan(&counterStruct)
	db.Database.Table("metrics").Select("name, value").Scopes(TypeIsGauge).Order("").Scan(&gaugeStruct)

	m.Counter = make(map[string]int64)
	m.Gauge = make(map[string]float64)
	for _, entry := range counterStruct {
		m.Counter[entry.Name] = entry.Delta
	}
	for _, entry := range gaugeStruct {
		m.Gauge[entry.Name] = entry.Value
	}

	return m.Counter, m.Gauge
}

// typeCondition is a helper function that returns a GORM scope function
// based on the metric type.
//
// Parameters:
// - param: A formatter.Metric struct used to determine the metric type.
//
// Returns:
// - A GORM scope function for the specific metric type.
func typeCondition(param formatter.Metric) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if param.MType == config.Counter {
			return db.Where("type = ?", config.Counter)
		} else if param.MType == config.Gauge {
			return db.Where("type = ?", config.Gauge)
		}
		return db
	}
}

// InsertBatchMetrics inserts multiple metrics into the database in a batch.
//
// Parameters:
// - metrics: A slice of formatter.Metric containing the metrics to be inserted.
//
// Returns:
// - An error if the batch insert or update operation fails.
func (db DB) InsertBatchMetrics(metrics []formatter.Metric) error {
	tx := db.Database.Begin()
	defer tx.Rollback()

	for _, param := range metrics {
		var m Metrics
		result := tx.Scopes(typeCondition(param)).Where("name = ?", param.ID).Order("").First(&m)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			var data *gorm.DB
			switch param.MType {
			case config.Counter:
				data = tx.Create(
					&Metrics{
						Name:  param.ID,
						Type:  config.Counter,
						Delta: *param.Delta,
					},
				)
			case config.Gauge:
				data = tx.Create(
					&Metrics{
						Name:  param.ID,
						Type:  config.Gauge,
						Value: *param.Value,
					},
				)
			}
			if data.Error != nil {
				logger.Log.Error(fmt.Sprintf("%s: %v", ErrCreateMetric, data.Error))
			}
		} else {
			var result *gorm.DB
			switch param.MType {
			case config.Counter:
				m.Delta += *param.Delta
			case config.Gauge:
				m.Value = *param.Value
			}
			result = tx.Save(&m)
			if result.Error != nil {
				logger.Log.Error(fmt.Sprintf("%s: %v", ErrSaveMetric, result.Error))
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("%s: %v", ErrCommitTransaction, err)
	}
	return nil
}
