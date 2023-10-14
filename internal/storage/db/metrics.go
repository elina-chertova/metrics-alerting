package db

import (
	"errors"
	"fmt"
	s "github.com/elina-chertova/metrics-alerting.git/internal/config"
	f "github.com/elina-chertova/metrics-alerting.git/internal/formatter"
	fm "github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"
	"gorm.io/gorm"
)

var (
	ErrRetrieveMetric    = errors.New("failed to retrieve metric")
	ErrSaveMetric        = errors.New("failed to save metric")
	ErrCreateMetric      = errors.New("failed to create metric")
	ErrCommitTransaction = errors.New("transaction commit error")
)

func TypeIsCounter(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", s.Counter)
}

func TypeIsGauge(db *gorm.DB) *gorm.DB {
	return db.Where("type = ?", s.Gauge)
}

func (db DB) UpdateCounter(name string, value int64, ok bool) {
	var m Metrics
	if ok {
		result := db.db.Scopes(TypeIsCounter).Where("name = ?", name).Order("").First(&m)
		if result.Error != nil {
			fmt.Printf("%s: %v\n", ErrRetrieveMetric, result.Error)
		}

		m.Delta += value

		result = db.db.Save(&m)
		if result.Error != nil {
			if result.Error != nil {
				fmt.Printf("%s: %v\n", ErrSaveMetric, result.Error)
			}
		}
		return
	}
	data := db.db.Create(
		&Metrics{
			Name:  name,
			Type:  s.Counter,
			Delta: value,
		},
	)
	if data.Error != nil {
		fmt.Printf("%s: %v\n", ErrCreateMetric, data.Error)
	}
}

func (db DB) UpdateGauge(name string, value float64) {
	var m Metrics

	if result := db.db.Scopes(TypeIsGauge).Where(
		"name = ?",
		name,
	).Order("").First(&m); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		data := db.db.Create(
			&Metrics{
				Name:  name,
				Type:  s.Gauge,
				Value: value,
			},
		)
		if data.Error != nil {
			fmt.Printf("%s: %v\n", ErrCreateMetric, data.Error)
		}
		return
	}
	m.Value = value

	result := db.db.Save(&m)
	if result.Error != nil {
		fmt.Printf("%s: %v\n", ErrSaveMetric, result.Error)
	}
}

func (db DB) GetCounter(name string) (int64, bool) {
	var m Metrics
	result := db.db.Scopes(TypeIsCounter).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false
	}
	return m.Delta, true
}

func (db DB) GetGauge(name string) (float64, bool) {
	var m Metrics
	result := db.db.Scopes(TypeIsGauge).Where("name = ?", name).Order("").First(&m)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, false
	}
	return m.Value, true
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

	var m fm.MemStorage
	db.db.Table("metrics").Select("name, delta").Scopes(TypeIsCounter).Order("").Scan(&counterStruct)
	db.db.Table("metrics").Select("name, value").Scopes(TypeIsGauge).Order("").Scan(&gaugeStruct)

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

func typeCondition(param f.Metric) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if param.MType == s.Counter {
			return db.Where("type = ?", s.Counter)
		} else if param.MType == s.Gauge {
			return db.Where("type = ?", s.Gauge)
		}
		return db
	}
}

func (db DB) InsertBatchMetrics(metrics []f.Metric) error {
	tx := db.db.Begin()
	defer tx.Rollback()

	for _, param := range metrics {
		var m Metrics
		result := tx.Scopes(typeCondition(param)).Where("name = ?", param.ID).Order("").First(&m)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			var data *gorm.DB
			switch param.MType {
			case s.Counter:
				data = tx.Create(
					&Metrics{
						Name:  param.ID,
						Type:  s.Counter,
						Delta: *param.Delta,
					},
				)
			case s.Gauge:
				data = tx.Create(
					&Metrics{
						Name:  param.ID,
						Type:  s.Gauge,
						Value: *param.Value,
					},
				)
			}
			if data.Error != nil {
				fmt.Printf("%s: %v\n", ErrCreateMetric, data.Error)
			}
		} else {
			var result *gorm.DB
			switch param.MType {
			case s.Counter:
				m.Delta += *param.Delta
			case s.Gauge:
				m.Value = *param.Value
			}
			result = tx.Save(&m)
			if result.Error != nil {
				fmt.Printf("%s: %v\n", ErrSaveMetric, result.Error)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("%s: %v\n", ErrCommitTransaction, err)
	}
	return nil
}
