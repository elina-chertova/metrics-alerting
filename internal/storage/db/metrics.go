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

func TypeIsCounter(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", config.Counter)
}

func TypeIsGauge(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", config.Gauge)
}

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

func (db DB) GetCounter(name string) (int64, bool, error) {
	var m Metrics
	result := db.Database.Scopes(TypeIsCounter).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	return m.Delta, true, nil
}

func (db DB) GetGauge(name string) (float64, bool, error) {
	var m Metrics
	result := db.Database.Scopes(TypeIsGauge).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false, nil
	}
	return m.Value, true, nil
}

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
