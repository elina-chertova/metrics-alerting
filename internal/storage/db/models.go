package db

import "github.com/jinzhu/gorm"

type Metrics struct {
	gorm.Model
	Name  string `gorm:"unique_index"`
	Type  string
	Delta int64
	Value float64 `gorm:"type:double precision"`
}
