// Auth: pp
// Created: 2025-03-21 0:17
// Description: the operation of database

package utils

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type DBConfig struct {
	Host 	string
	Port 	string
	User 	string
	Passwd 	string
	DBName 	string
}

// create a new database connection and return it
func NewDBConnection(cfg *DBConfig) (*gorm.DB, error) {
	// get the database configuration
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Passwd, cfg.Host, cfg.Port, cfg.DBName)
	// create a new database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

// create a new database connection pool and return it
// @param db *gorm.DB: the database connection
// @param maxOpenConns int: the maximum number of open connections to the database
// @param maxIdleConns int: the maximum number of connections in the idle connection pool
// @param connMaxLifetime int: the maximum amount of time a connection may be reused
func SetupDBConnectionPool(db *gorm.DB, maxOpenConns, maxIdleConns, connMaxLifetime int) error {
	if maxOpenConns <= 0 || maxIdleConns < 0 || connMaxLifetime <= 0 {
		return fmt.Errorf("invalid connection pool parameters: maxOpen=%d, maxIdle=%d, maxLifetime=%v",
			maxOpenConns, maxIdleConns, connMaxLifetime)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	log.Printf("Database connection pool configured: maxOpen=%d, maxIdle=%d, maxLifetime=%v", 
		maxOpenConns, maxIdleConns, connMaxLifetime)
	return nil
}

// show the database connection pool status immediately
// @param db *gorm.DB: the database connection
func ShowPoolStatus(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}

	stats := sqlDB.Stats()
	log.Printf("Database connection pool status: %+v", stats)

	// 打印关键指标
	log.Printf("Open Connections: %d", stats.OpenConnections)
	log.Printf("In-use Connections: %d", stats.InUse)
	log.Printf("Idle Connections: %d", stats.Idle)
	log.Printf("Wait Count: %d", stats.WaitCount)
	log.Printf("Wait Duration: %v", stats.WaitDuration)

	return nil
}

// generate a new UUID and return it
func GenerateUUID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return uuid.String(), nil
}