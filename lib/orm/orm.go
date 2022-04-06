package orm

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"g.hz.netease.com/horizon/lib/q"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
)

// MySQL ...
type MySQL struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	Database          string `json:"database"`
	PrometheusEnabled bool   `json:"prometheusEnabled"`
}

func NewMySQLDB(db *MySQL) (*gorm.DB, error) {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.Username,
		db.Password, db.Host, db.Port, db.Database)

	sqlDB, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	orm, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
	})

	if db.PrometheusEnabled {
		if err := orm.Use(prometheus.New(prometheus.Config{
			DBName: "mysql",
			MetricsCollector: []prometheus.MetricsCollector{
				&MySQLMetricsCollector{},
			},
		})); err != nil {
			return nil, err
		}
	}

	return orm, err
}

func NewSqliteDB(file string) (*gorm.DB, error) {
	orm, err := gorm.Open(sqlite.Open(file), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             0, // print all logs
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		),
	})

	return orm, err
}

func FormatSortExp(query *q.Query) string {
	exp := ""

	if query == nil || query.Sorts == nil || len(query.Sorts) == 0 {
		return exp
	}

	for _, sort := range query.Sorts {
		sorting := sort.Key
		if sort.DESC {
			sorting = fmt.Sprintf("%s desc", sorting)
		}
		exp += sorting + ", "
	}

	return exp[:len(exp)-2]
}

// FormatFilterExp returns a where condition string which has prefixed "and"
func FormatFilterExp(query *q.Query, columnInTable map[string]string) (string, []interface{}) {
	if query == nil || query.Keywords == nil || len(query.Keywords) == 0 {
		return "", []interface{}{}
	}

	exp := strings.Builder{}
	values := make([]interface{}, 0, len(query.Keywords))

	for filterKey, filterValue := range query.Keywords {
		value, ok := filterValue.(string)
		if filterKey == "" || !ok || value == "" {
			continue
		}
		if columnInTable != nil {
			if keyInDB, ok := columnInTable[filterKey]; ok {
				filterKey = keyInDB
			}
		}
		exp.WriteString(fmt.Sprintf(" %s = ? and ", filterKey))
		values = append(values, value)
	}
	return exp.String(), values
}
