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
	log.Printf("%s: %v", message, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
