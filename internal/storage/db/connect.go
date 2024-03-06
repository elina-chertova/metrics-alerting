package db

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type DB struct {
	db *gorm.DB
}

func Connect(dsn string) *DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database because %s", err)
	}
	db.AutoMigrate(&Metrics{})
	return &DB{db: db}
}

func (db *DB) PingDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.db.DB()
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database connection error"},
			)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"message": "Database connection error"},
			)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{"message": "Successfully connected to the database and pinged it"},
		)
	}
}
