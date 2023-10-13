package db

import (
	"fmt"
	s "github.com/elina-chertova/metrics-alerting.git/internal/storage"
	fm "github.com/elina-chertova/metrics-alerting.git/internal/storage/file_memory"
)

func (db *DB) UpdateCounter(name string, value int64, ok bool) {
	var m Metrics
	if ok {
		result := db.db.Where("name = ? and type = ?", name, s.Counter).Order("").First(&m)
		if result.Error != nil {
			_ = fmt.Errorf("failed to retrieve metric: " + result.Error.Error())
		}

		m.Delta += value

		result = db.db.Save(&m)
		if result.Error != nil {
			_ = fmt.Errorf("failed to save metric: " + result.Error.Error())
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
		_ = fmt.Errorf("failed to create metric: " + data.Error.Error())
	}
}

func (db *DB) UpdateGauge(name string, value float64) {
	var m Metrics

	if result := db.db.Where(
		"name = ? and type = ?",
		name,
		s.Gauge,
	).Order("").First(&m); result.Error != nil {
		data := db.db.Create(
			&Metrics{
				Name:  name,
				Type:  s.Gauge,
				Value: value,
			},
		)
		if data.Error != nil {
			_ = fmt.Errorf("failed to create metric: " + data.Error.Error())
		}
		return
	}
	m.Value = value

	result := db.db.Save(&m)
	if result.Error != nil {
		_ = fmt.Errorf("failed to save metric: " + result.Error.Error())
	}
}

func (db *DB) GetCounter(name string) (int64, bool) {
	var m Metrics
	result := db.db.Where("name = ? and type = ?", name, s.Counter).Order("").First(&m)
	if result.Error != nil {
		return 0, false
	}
	return m.Delta, true
}

func (db *DB) GetGauge(name string) (float64, bool) {
	var m Metrics
	result := db.db.Where("name = ? and type = ?", name, s.Gauge).Order("").First(&m)
	if result.Error != nil {
		return 0, false
	}
	return m.Value, true
}
func (db *DB) GetMetrics() (map[string]int64, map[string]float64) {
	var counterStruct []struct {
		Name  string
		Delta int64
	}
	var gaugeStruct []struct {
		Name  string
		Value float64
	}

	var m fm.MemStorage
	db.db.Table("metrics").Select("name, delta").Where(
		"type = ?",
		s.Counter,
	).Order("").Scan(&counterStruct)
	db.db.Table("metrics").Select("name, value").Where(
		"type = ?",
		s.Gauge,
	).Order("").Scan(&gaugeStruct)

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
