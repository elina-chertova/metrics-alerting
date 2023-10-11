package db

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
)

func Connect(url string) *sql.DB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatalf("Unable to connect to database because %s", err)
	}
	return db
}

func PingDB(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Ping(); err != nil {
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
