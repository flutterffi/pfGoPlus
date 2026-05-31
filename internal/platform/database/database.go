package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	if cfg.Driver != "sqlite" {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.DSN), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	log.Info("database connected", zap.String("driver", cfg.Driver), zap.String("dsn", cfg.DSN))
	return db, nil
}
