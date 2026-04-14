package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB opens a GORM connection with retry, configures the pool, and auto-migrates all models.
func InitDB(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Retry connection up to 30 times (covers slow MySQL startup in Docker)
	for i := 0; i < 30; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				pingErr := sqlDB.PingContext(ctx)
				cancel()
				if pingErr == nil {
					break
				}
				err = pingErr
			} else {
				err = dbErr
			}
		}
		log.Printf("db connect attempt %d/30 failed: %v — retrying in 2s", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("open db after retries: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate all domain models
	if err := db.AutoMigrate(
		&User{},
		&Plot{},
		&Device{},
		&Metric{},
		&Task{},
		&AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	return db, nil
}

// Ping checks database connectivity with a context timeout.
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}
