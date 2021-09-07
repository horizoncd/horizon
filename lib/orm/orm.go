package main

import (
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DSN string

var db *gorm.DB

func main() {
	// TODO remove
	_ = os.Setenv("DATABASE_URL", "horizon:O{AyroRA@@tcp(music-horizon-dev-jd-26519.rds.cn-east-p1.internal:3331)/horizon?charset=utf8mb4&parseTime=True&loc=Local")
	DSN = os.Getenv("DATABASE_URL")
	db, _ = gorm.Open(mysql.New(mysql.Config{
		DSN:                       DSN,
		SkipInitializeWithVersion: false,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
	}), &gorm.Config{
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	})

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	type User struct {
		gorm.Model
		Name string
		Age int
		Birthday time.Time
	}

	// m := db.Migrator()
	// _ = m.CreateTable(&User{})

	db.Create(&User{
		Name:     "wurongjun",
		Age:      20,
		Birthday: time.Now(),
	})

}

