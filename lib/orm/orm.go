// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import (
	"database/sql"
	"fmt"
	"github.com/horizoncd/horizon/pkg/config/db"
	"log"
	"os"
	"strings"
	"time"

	"github.com/horizoncd/horizon/lib/q"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
)

func NewMySQLDB(db db.Config) (*gorm.DB, error) {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.Username,
		db.Password, db.Host, db.Port, db.Database)

	// Set default value for SlowThreshold if not provided
	if db.SlowThreshold == 0 {
		db.SlowThreshold = 200 * time.Millisecond
	}
	// Set default value for MaxIdleConns if not provided
	if db.MaxIdleConns == 0 {
		db.MaxIdleConns = 10
	}
	// Set default value for MaxOpenConns if not provided
	if db.MaxOpenConns == 0 {
		db.MaxOpenConns = 100
	}

	sqlDB, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(db.MaxIdleConns)
	sqlDB.SetMaxOpenConns(db.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	orm, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             db.SlowThreshold,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			}),
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
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
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

func ValidateQuery(query q.Query, columnInTable map[string]string) map[string]interface{} {
	if query.Keywords == nil || len(query.Keywords) == 0 {
		return nil
	}

	res := make(map[string]interface{}, len(query.Keywords))
	for filterKey, filterValue := range query.Keywords {
		if mappingKey, ok := columnInTable[filterKey]; ok {
			res[mappingKey] = filterValue
		}
	}
	return res
}
