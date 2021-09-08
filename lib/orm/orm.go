package orm

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm/schema"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQL ...
type MySQL struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Database string `json:"database"`
}

func NewMySQLDB(db *MySQL) (*gorm.DB, error) {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.Username,
		db.Password, db.Host, db.Port, db.Database)

	sqlDB, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	orm, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	return orm, err
}
