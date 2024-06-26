package db

import (
	"fmt"
	"log"
	"net/http"

	"github.com/elina-chertova/metrics-alerting.git/internal/middleware/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Database *gorm.DB
}

func Connect(dsn string) *DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database because %s", err)
	}
	db.AutoMigrate(&Metrics{})
	return &DB{Database: db}
}

func (db *DB) PingDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.Database.DB()
		if err != nil {
			handleDBError(c, "failed to get database connection", err)
			return
		}

		if err := sqlDB.Ping(); err != nil {
			handleDBError(c, "failed to ping the database", err)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{"message": "Successfully connected to the database and pinged it"},
		)
	}
}

func handleDBError(c *gin.Context, message string, err error) {
	logger.Log.Error(fmt.Sprintf("%s: %v", message, err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
