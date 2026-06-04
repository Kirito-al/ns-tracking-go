package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB PostgreSQL 数据库连接封装（GORM）
// 对标：demo1-gozero service/tracking/rpc/internal/svc/servicecontext.go
type DB struct {
	*gorm.DB
}

// NewDB 创建 GORM 数据库连接（连接池配置）
// 对标：demo1-gozero service/tracking/rpc/internal/svc/servicecontext.go:43-70
func NewDB(dataSource string) (*DB, error) {
	// GORM 配置（Silent 日志模式）
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(dataSource), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// 连接池配置（对标 Ruby Rails connection pool）
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	return &DB{db}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}