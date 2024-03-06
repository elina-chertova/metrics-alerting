package db

//
//import (
//	"testing"
//
//	"github.com/DATA-DOG/go-sqlmock"
//	"gorm.io/driver/postgres"
//	"gorm.io/gorm"
//)
//
//func setupMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {
//	mockDB, mock, err := sqlmock.New() // Mock SQL database
//	if err != nil {
//		return nil, nil, err
//	}
//
//	dialector := postgres.New(postgres.Config{
//		Conn: mockDB,
//	})
//
//	db, err := gorm.Open(dialector, &gorm.Config{})
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return db, mock, nil
//}
//
//func TestUpdateCounter(t *testing.T) {
//	db, mock, err := setupMockDB()
//	if err != nil {
//		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
//	}
//	defer db.Close()
//
//	// Define your test cases here
//	// For example:
//	t.Run("increment existing counter", func(t *testing.T) {
//		// Set up your mock expectations
//		// e.g., mock.ExpectQuery(...).WillReturnRows(...)
//
//		// Call your function
//		// e.g., err := db.UpdateCounter("testMetric", 1, true)
//
//		// Assert expectations and results
//		// e.g., assert.NoError(t, err)
//	})
//
//	// Remember to verify that all expectations were met
//	if err := mock.ExpectationsWereMet(); err != nil {
//		t.Errorf("there were unfulfilled expectations: %s", err)
//	}
//}
//
//// Similarly, write tests for UpdateGauge, GetCounter, GetGauge, GetMetrics, and InsertBatchMetrics
